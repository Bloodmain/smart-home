package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool,
	}
}

const saveEventQuery = `insert into db.public.events (timestamp, sensor_serial_number, sensor_id, payload)
	values ($1, $2, $3, $4);`

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	_, err := r.pool.Exec(ctx, saveEventQuery, event.Timestamp, event.SensorSerialNumber, event.SensorID, event.Payload)
	if err != nil {
		return err
	}
	return ctx.Err()
}

const getLastEventBySensorIDQuery = `
select *
from db.public.events 
where sensor_id=$1
order by timestamp desc;`

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx, getLastEventBySensorIDQuery, id)

	event := &domain.Event{}
	if err := row.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrEventNotFound
		}
		return nil, fmt.Errorf("can't scan event: %w", err)
	}

	return event, ctx.Err()
}
