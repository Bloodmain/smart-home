package usecase

import (
	"context"
	"homework/internal/domain"
)

type Event struct {
	er EventRepository
	sr SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{er: er, sr: sr}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}
	s, err := e.sr.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return err
	}

	event.SensorID = s.ID
	s.LastActivity = event.Timestamp
	s.CurrentState = event.Payload

	if err = e.er.SaveEvent(ctx, event); err != nil {
		return err
	}
	return e.sr.SaveSensor(ctx, s)
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if _, err := e.sr.GetSensorByID(ctx, id); err != nil {
		return nil, err
	}
	return e.er.GetLastEventBySensorID(ctx, id)
}
