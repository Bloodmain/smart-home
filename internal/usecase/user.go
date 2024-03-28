package usecase

import (
	"context"
	"homework/internal/domain"
)

type User struct {
	ur  UserRepository
	sr  SensorRepository
	sor SensorOwnerRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{ur: ur, sr: sr, sor: sor}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if len(user.Name) == 0 {
		return nil, ErrInvalidUserName
	}
	return user, u.ur.SaveUser(ctx, user)
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	if _, err := u.ur.GetUserByID(ctx, userID); err != nil {
		return err
	}
	if _, err := u.sr.GetSensorByID(ctx, sensorID); err != nil {
		return err
	}
	return u.sor.SaveSensorOwner(ctx, domain.SensorOwner{UserID: userID, SensorID: sensorID})
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	if _, err := u.ur.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	sOwners, err := u.sor.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	done := make(chan struct{})
	e := make(chan error)
	s := make([]domain.Sensor, 0, len(sOwners))

	go func() {
		for _, so := range sOwners {
			sensor, err := u.sr.GetSensorByID(ctx, so.SensorID)
			if err != nil {
				e <- err
				return
			}
			s = append(s, *sensor)
		}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err = <-e:
		return nil, err
	case <-done:
		return s, nil
	}
}
