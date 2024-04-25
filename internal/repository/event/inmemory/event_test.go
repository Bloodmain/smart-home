package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	t.Run("ok, save and get all events", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		size := 10
		originalEvents := make([]*domain.Event, 0, size)

		for i := 0; i < size; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   int64(i),
			}
			originalEvents = append(originalEvents, event)
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		events, err := er.GetHistoryBySensorID(ctx, sensorID, time.Time{}, time.Now())
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, size, len(events))

		for i, event := range events {
			assert.Equal(t, originalEvents[i], event)
		}
	})

	t.Run("ok, save and get segment including bounds", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		size := 10
		originalEvents := make([]*domain.Event, 0, size)

		for i := 0; i < size; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   int64(i),
			}
			originalEvents = append(originalEvents, event)
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		events, err := er.GetHistoryBySensorID(ctx, sensorID, originalEvents[1].Timestamp, originalEvents[5].Timestamp)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, 5, len(events))

		for i, event := range events {
			assert.Equal(t, originalEvents[1+i], event)
		}
	})

	t.Run("ok, save and get segment excluding bounds", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		size := 10
		originalEvents := make([]*domain.Event, 0, size)

		for i := 0; i < size; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   int64(i),
			}
			originalEvents = append(originalEvents, event)
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		from := originalEvents[1].Timestamp.Add(5 * time.Millisecond)
		to := originalEvents[5].Timestamp.Add(-5 * time.Millisecond)
		events, err := er.GetHistoryBySensorID(ctx, sensorID, from, to)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, 3, len(events))

		for i, event := range events {
			assert.Equal(t, originalEvents[2+i], event)
		}
	})

	t.Run("ok, from = to", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		size := 10
		originalEvents := make([]*domain.Event, 0, size)

		for i := 0; i < size; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   int64(i),
			}
			originalEvents = append(originalEvents, event)
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		events, err := er.GetHistoryBySensorID(ctx, sensorID, originalEvents[8].Timestamp, originalEvents[8].Timestamp)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, originalEvents[8], events[0])
	})

	t.Run("ok, empty", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  1,
				Payload:   int64(i),
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		events, err := er.GetHistoryBySensorID(ctx, 1, time.Now(), time.Now().Add(10*time.Second))
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, 0, len(events))
	})

	t.Run("ok, from > to", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  1,
				Payload:   int64(i),
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		events, err := er.GetHistoryBySensorID(ctx, 1, time.Now(), time.Time{})
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, 0, len(events))
	})
}
