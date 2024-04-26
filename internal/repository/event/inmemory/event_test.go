package inmemory

import (
	"context"
	"github.com/stretchr/testify/assert"
	"homework/internal/domain"
	"homework/internal/usecase"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

func TestEventRepository_SaveEvent(t *testing.T) {
	t.Run("err, event is nil", func(t *testing.T) {
		er := NewEventRepository()
		err := er.SaveEvent(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		event := &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "12345",
			Payload:            0,
		}

		err := er.SaveEvent(ctx, event)
		assert.NoError(t, err)

		actualEvent, err := er.GetLastEventBySensorID(ctx, event.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, event.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, event.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, event.Payload, actualEvent.Payload)
	})

	t.Run("ok, collision test", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := sync.WaitGroup{}
		var lastEvent domain.Event
		for i := 0; i < 1000; i++ {
			event := &domain.Event{
				Timestamp:          time.Now(),
				SensorSerialNumber: "12345",
				Payload:            0,
			}
			lastEvent = *event
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, er.SaveEvent(ctx, event))
			}()
		}

		wg.Wait()

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetLastEventBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, event not found", func(t *testing.T) {
		er := NewEventRepository()
		_, err := er.GetLastEventBySensorID(context.Background(), 234)
		assert.ErrorIs(t, err, usecase.ErrEventNotFound)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		var lastEvent *domain.Event
		for i := 0; i < 10; i++ {
			lastEvent = &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   0,
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, lastEvent))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetHistoryBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetHistoryBySensorID(ctx, 0, time.Now(), time.Now())
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetHistoryBySensorID(ctx, 0, time.Now(), time.Now())
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, event not found", func(t *testing.T) {
		er := NewEventRepository()
		_, err := er.GetHistoryBySensorID(context.Background(), 456, time.Now(), time.Now())
		assert.ErrorIs(t, err, usecase.ErrEventNotFound)
	})

	t.Run("ok, validate results", func(t *testing.T) {
		tests := []struct {
			name  string
			n     int
			query func(ctx context.Context, er *EventRepository, originalEvents []*domain.Event) ([]*domain.Event, error)
			check func(t *testing.T, events []*domain.Event, err error, originalEvents []*domain.Event)
		}{
			{
				name: "all events",
				n:    10,
				query: func(ctx context.Context, er *EventRepository, _ []*domain.Event) ([]*domain.Event, error) {
					return er.GetHistoryBySensorID(ctx, 12345, time.Time{}, time.Now())
				},
				check: func(t *testing.T, events []*domain.Event, err error, originalEvents []*domain.Event) {
					assert.NoError(t, err)
					assert.NotNil(t, events)
					assert.Equal(t, len(originalEvents), len(events))

					for i, event := range events {
						assert.Equal(t, originalEvents[i], event)
					}
				},
			},
			{
				name: "segment including bounds",
				n:    10,
				query: func(ctx context.Context, er *EventRepository, originalEvents []*domain.Event) ([]*domain.Event, error) {
					return er.GetHistoryBySensorID(ctx, 12345, originalEvents[1].Timestamp, originalEvents[5].Timestamp)
				},
				check: func(t *testing.T, events []*domain.Event, err error, originalEvents []*domain.Event) {
					assert.NoError(t, err)
					assert.NotNil(t, events)
					assert.Equal(t, 5, len(events))

					for i, event := range events {
						assert.Equal(t, originalEvents[1+i], event)
					}
				},
			},
			{
				name: "segment excluding bounds",
				n:    10,
				query: func(ctx context.Context, er *EventRepository, originalEvents []*domain.Event) ([]*domain.Event, error) {
					from := originalEvents[1].Timestamp.Add(5 * time.Millisecond)
					to := originalEvents[5].Timestamp.Add(-5 * time.Millisecond)
					return er.GetHistoryBySensorID(ctx, 12345, from, to)
				},
				check: func(t *testing.T, events []*domain.Event, err error, originalEvents []*domain.Event) {
					assert.NoError(t, err)
					assert.NotNil(t, events)
					assert.Equal(t, 3, len(events))

					for i, event := range events {
						assert.Equal(t, originalEvents[2+i], event)
					}
				},
			},
			{
				name: "from = to",
				n:    10,
				query: func(ctx context.Context, er *EventRepository, originalEvents []*domain.Event) ([]*domain.Event, error) {
					return er.GetHistoryBySensorID(ctx, 12345, originalEvents[8].Timestamp, originalEvents[8].Timestamp)
				},
				check: func(t *testing.T, events []*domain.Event, err error, originalEvents []*domain.Event) {
					assert.NoError(t, err)
					assert.NotNil(t, events)
					assert.Equal(t, 1, len(events))
					assert.Equal(t, originalEvents[8], events[0])
				},
			},
			{
				name: "empty",
				n:    10,
				query: func(ctx context.Context, er *EventRepository, _ []*domain.Event) ([]*domain.Event, error) {
					return er.GetHistoryBySensorID(ctx, 12345, time.Now(), time.Now().Add(10*time.Second))
				},
				check: func(t *testing.T, events []*domain.Event, err error, _ []*domain.Event) {
					assert.NoError(t, err)
					assert.NotNil(t, events)
					assert.Equal(t, 0, len(events))
				},
			},
			{
				name: "from > to",
				n:    10,
				query: func(ctx context.Context, er *EventRepository, _ []*domain.Event) ([]*domain.Event, error) {
					return er.GetHistoryBySensorID(ctx, 12345, time.Now(), time.Time{})
				},
				check: func(t *testing.T, events []*domain.Event, err error, _ []*domain.Event) {
					assert.NoError(t, err)
					assert.NotNil(t, events)
					assert.Equal(t, 0, len(events))
				},
			},
		}

		setupEvents := func(n int, er *EventRepository) []*domain.Event {
			originalEvents := make([]*domain.Event, n)

			for i := 0; i < n; i++ {
				event := &domain.Event{
					Timestamp: time.Now(),
					SensorID:  12345,
					Payload:   int64(i),
				}
				originalEvents[i] = event
				time.Sleep(10 * time.Millisecond)
				assert.NoError(t, er.SaveEvent(context.Background(), event))
			}
			return originalEvents
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				er := NewEventRepository()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				original := setupEvents(tt.n, er)
				got, err := tt.query(ctx, er, original)
				tt.check(t, got, err, original)
			})
		}
	})
}

func FuzzEventRepository_GetHistoryBySensorID(f *testing.F) {
	f.Add(int64(223423), int64(95747433))

	n := 100000
	originalEvents := make([]*domain.Event, n)
	er := NewEventRepository()

	for i := 0; i < n; i++ {
		event := &domain.Event{
			Timestamp: time.Unix(rand.Int64(), 0),
			SensorID:  12345,
			Payload:   int64(i),
		}
		originalEvents[i] = event
		_ = er.SaveEvent(context.Background(), event)
	}

	f.Fuzz(func(t *testing.T, from, to int64) {
		events, err := er.GetHistoryBySensorID(context.Background(), 12345, time.Unix(from, 0), time.Unix(to, 0))

		assert.NoError(t, err)
		assert.NotNil(t, events)

		for _, event := range events {
			assert.LessOrEqual(t, from, event.Timestamp.Unix())
			assert.GreaterOrEqual(t, to, event.Timestamp.Unix())
		}

		for _, event := range originalEvents {
			if from <= event.Timestamp.Unix() && event.Timestamp.Unix() <= to {
				assert.Contains(t, events, event)
			}
		}
	})
}
