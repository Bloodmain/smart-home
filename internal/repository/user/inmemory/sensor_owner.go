package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
)

type SensorOwnerRepository struct {
	storage []domain.SensorOwner
	m       sync.RWMutex
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{storage: []domain.SensorOwner{}, m: sync.RWMutex{}}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	r.m.Lock()
	r.storage = append(r.storage, sensorOwner)
	r.m.Unlock()
	return ctx.Err()
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	done := make(chan struct{})
	sensors := make([]domain.SensorOwner, 0, len(r.storage))

	go func() {
		r.m.RLock()
	outer:
		for _, v := range r.storage {
			select {
			case <-ctx.Done():
				break outer
			default:
				if v.UserID == userID {
					sensors = append(sensors, v)
				}
			}
		}
		r.m.RUnlock()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return sensors, nil
	}
}
