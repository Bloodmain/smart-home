package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

var ErrNilSensorPointer = errors.New("nil sensor is provided")

type SensorSerialNumber string

type SensorRepository struct {
	storage map[SensorSerialNumber]*domain.Sensor
	m       sync.RWMutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{storage: map[SensorSerialNumber]*domain.Sensor{}, m: sync.RWMutex{}}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor == nil {
		return ErrNilSensorPointer
	}
	sensor.RegisteredAt = time.Now()
	r.m.Lock()
	r.storage[SensorSerialNumber(sensor.SerialNumber)] = sensor
	r.m.Unlock()
	return ctx.Err()
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	sensors := make([]domain.Sensor, 0, len(r.storage))

	r.m.RLock()
	defer r.m.RUnlock()
	for _, v := range r.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			sensors = append(sensors, *v)
		}
	}

	return sensors, ctx.Err()
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	for _, v := range r.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if v.ID == id {
				return v, ctx.Err()
			}
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, usecase.ErrSensorNotFound
	}
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	sensor, has := r.storage[SensorSerialNumber(sn)]
	if !has {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, usecase.ErrSensorNotFound
	}
	return sensor, ctx.Err()
}
