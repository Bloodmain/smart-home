package usecase

import (
	"context"
	"homework/internal/domain"
)

type Event struct {
	er EventRepository
	sr SensorRepository
	eventRepository  EventRepository
	sensorRepository SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{er: er, sr: sr}
	return &Event{eventRepository: er, sensorRepository: sr}
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
