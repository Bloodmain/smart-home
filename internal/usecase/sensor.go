package usecase

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/repository/sensor/inmemory"
	"regexp"
)

const (
	sensorSerialNumberLength = 10
)

type Sensor struct {
	sensorRepository SensorRepository
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{sensorRepository: sr}
}

func validate(sensor *domain.Sensor) error {
	if _, has := domain.AcceptableSensorTypes[sensor.Type]; !has {
		return ErrWrongSensorType
	}
	reg := regexp.MustCompile(fmt.Sprintf("^\\d{%d}$", sensorSerialNumberLength))
	if m := reg.MatchString(sensor.SerialNumber); !m {
		return ErrWrongSensorSerialNumber
	}
	return nil
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if err := validate(sensor); err != nil {
		return nil, err
	}
	old, err := s.sensorRepository.GetSensorBySerialNumber(ctx, sensor.SerialNumber)
	if err != nil {
		if errors.Is(err, inmemory.ErrSensorNotFound) {
			if err = s.sensorRepository.SaveSensor(ctx, sensor); err != nil {
				return nil, err
			}
			return sensor, nil
		}

		return nil, err
	}

	return old, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	return s.sensorRepository.GetSensors(ctx)
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	return s.sensorRepository.GetSensorByID(ctx, id)
}
