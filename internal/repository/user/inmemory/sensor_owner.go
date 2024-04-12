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
	sensors := make([]domain.SensorOwner, 0, len(r.storage))

	r.m.RLock()
	defer r.m.RUnlock()
	for _, v := range r.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if v.UserID == userID {
				sensors = append(sensors, v)
			}
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return sensors, nil
	}
}
