package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
)

var (
	ErrEventNotFound   = errors.New("event not found")
	ErrNilEventPointer = errors.New("nil event is provided")
)

type EventRepository struct {
	// maps sensor's id to its last event
	lastEvent map[int64]*domain.Event
	m         sync.RWMutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{lastEvent: map[int64]*domain.Event{}, m: sync.RWMutex{}}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return ErrNilEventPointer
	}
	r.m.Lock()
	now, has := r.lastEvent[event.SensorID]
	if !has || event.Timestamp.After(now.Timestamp) {
		r.lastEvent[event.SensorID] = event
	}
	r.m.Unlock()
	return ctx.Err()
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	r.m.RLock()
	event, has := r.lastEvent[id]
	r.m.RUnlock()
	if !has {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrEventNotFound
	}
	return event, ctx.Err()
}
