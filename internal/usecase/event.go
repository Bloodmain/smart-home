package usecase

import (
	"context"
	"homework/internal/domain"
	"time"
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
	return e.eventRepository.GetLastEventBySensorID(ctx, id)
}

func (e *Event) GetHistoryBySensorID(ctx context.Context, id int64, from, to time.Time) ([]*domain.Event, error) {
	return e.eventRepository.GetHistoryBySensorID(ctx, id, from, to)
}
