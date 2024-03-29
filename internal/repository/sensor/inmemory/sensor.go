package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
	"time"
)

var (
	ErrSensorNotFound   = errors.New("sensor not found")
	ErrNilSensorPointer = errors.New("nil sensor is provided")
)

type SensorRepository struct {
	storage []*domain.Sensor
	m       sync.RWMutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{storage: []*domain.Sensor{}, m: sync.RWMutex{}}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor == nil {
		return ErrNilSensorPointer
	}
	sensor.RegisteredAt = time.Now()
	r.m.Lock()
	r.storage = append(r.storage, sensor)
	r.m.Unlock()
	return ctx.Err()
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	sensors := make([]domain.Sensor, 0, len(r.storage))

	r.m.RLock()
	for _, v := range r.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			sensors = append(sensors, *v)
		}
	}
	r.m.RUnlock()

	return sensors, ctx.Err()
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	return r.getSensorFunc(ctx, func(sensor *domain.Sensor) bool {
		return sensor.ID == id
	})
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	return r.getSensorFunc(ctx, func(sensor *domain.Sensor) bool {
		return sensor.SerialNumber == sn
	})
}

func (r *SensorRepository) getSensorFunc(ctx context.Context, p func(sensor *domain.Sensor) bool) (*domain.Sensor, error) {
	r.m.RLock()
	for _, v := range r.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if p(v) {
				return v, ctx.Err()
			}
		}
	}
	r.m.RUnlock()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, ErrSensorNotFound
	}
}
