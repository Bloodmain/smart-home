package inmemory

import (
	"context"
	"errors"
	"github.com/emirpasic/gods/trees/redblacktree"
	"homework/internal/domain"
	"homework/internal/usecase"
	"slices"
	"sync"
	"time"
)

var ErrNilEventPointer = errors.New("nil event is provided")

type SensorId int64

type EventRepository struct {
	// maps sensor to all of its event compared by timestamps
	events map[SensorId]*redblacktree.Tree
	m      sync.RWMutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{events: map[SensorId]*redblacktree.Tree{}, m: sync.RWMutex{}}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return ErrNilEventPointer
	}
	r.m.Lock()
	defer r.m.Unlock()

	tree, has := r.events[SensorId(event.SensorID)]
	if !has {
		tree = redblacktree.NewWith(func(a, b interface{}) int {
			s1 := a.(time.Time)
			s2 := b.(time.Time)
			return s1.Compare(s2)
		})
		r.events[SensorId(event.SensorID)] = tree
	}
	tree.Put(event.Timestamp, *event)

	return ctx.Err()
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	tree, has := r.events[SensorId(id)]
	if !has {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, usecase.ErrEventNotFound
	}
	it := tree.Iterator()
	it.Last()
	v := it.Value().(domain.Event)
	return &v, ctx.Err()
}

func (r *EventRepository) GetHistoryBySensorID(ctx context.Context, id int64, from, to time.Time) ([]*domain.Event, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	tree, has := r.events[SensorId(id)]
	if !has {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, usecase.ErrEventNotFound
	}

	v := tree.Values()
	comparator := func(e interface{}, t time.Time) int {
		return e.(domain.Event).Timestamp.Compare(t)
	}
	lb, _ := slices.BinarySearchFunc(v, from, comparator)
	rb, found := slices.BinarySearchFunc(v, to, comparator)

	// result size, excluding rb-th element
	resSize := rb - lb
	// if rb is actually found we should include it in our result
	if found {
		resSize++
	}

	if resSize <= 0 {
		return []*domain.Event{}, nil
	}

	res := make([]*domain.Event, 0, resSize)
	for i := lb; i < lb+resSize; i++ {
		e := v[i].(domain.Event)
		res = append(res, &e)
	}

	return res, ctx.Err()
}
