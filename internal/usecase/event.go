package usecase

import (
	"context"
	"homework/internal/domain"
)

type Event struct {
	eventRepository  EventRepository
	sensorRepository SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{eventRepository: er, sensorRepository: sr}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}
	s, err := e.sensorRepository.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return err
	}

	event.SensorID = s.ID
	s.LastActivity = event.Timestamp
	s.CurrentState = event.Payload

	if err = e.eventRepository.SaveEvent(ctx, event); err != nil {
		return err
	}
	return e.sensorRepository.SaveSensor(ctx, s)
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if _, err := e.sensorRepository.GetSensorByID(ctx, id); err != nil {
		return nil, err
	}
	return e.eventRepository.GetLastEventBySensorID(ctx, id)
}
