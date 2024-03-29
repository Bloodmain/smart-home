package usecase

import (
	"context"
	"homework/internal/domain"
)

type User struct {
	userRepository        UserRepository
	sensorRepository      SensorRepository
	sensorOwnerRepository SensorOwnerRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{userRepository: ur, sensorRepository: sr, sensorOwnerRepository: sor}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if len(user.Name) == 0 {
		return nil, ErrInvalidUserName
	}
	return user, u.userRepository.SaveUser(ctx, user)
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	if _, err := u.userRepository.GetUserByID(ctx, userID); err != nil {
		return err
	}
	if _, err := u.sensorRepository.GetSensorByID(ctx, sensorID); err != nil {
		return err
	}
	return u.sensorOwnerRepository.SaveSensorOwner(ctx, domain.SensorOwner{UserID: userID, SensorID: sensorID})
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	if _, err := u.userRepository.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	sOwners, err := u.sensorOwnerRepository.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	done := make(chan struct{})
	e := make(chan error)
	s := make([]domain.Sensor, 0, len(sOwners))

	go func() {
	outer:
		for _, so := range sOwners {
			select {
			case <-ctx.Done():
				break outer
			default:
				sensor, err := u.sensorRepository.GetSensorByID(ctx, so.SensorID)
				if err != nil {
					e <- err
					return
				}
				s = append(s, *sensor)
			}
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
