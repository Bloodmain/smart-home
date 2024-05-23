package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

const saveSensorQuery = `
insert into db.public.sensors (serial_number, type, current_state, description, is_active, registered_at, last_activity) 
values ($1, $2, $3, $4, $5, $6, $7)`

const updateSensorQuery = `
update db.public.sensors 
set current_state = $2, description = $3, is_active = $4, last_activity = $5
where serial_number = $1`

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	var err error
	if _, e := r.GetSensorBySerialNumber(ctx, sensor.SerialNumber); e == nil {
		_, err = r.pool.Exec(ctx, updateSensorQuery, sensor.SerialNumber, sensor.CurrentState, sensor.Description, sensor.IsActive, sensor.LastActivity)
	} else {
		sensor.RegisteredAt = time.Now()
		_, err = r.pool.Exec(ctx, saveSensorQuery, sensor.SerialNumber, sensor.Type, sensor.CurrentState, sensor.Description, sensor.IsActive, sensor.RegisteredAt, sensor.LastActivity)
	}

	if err != nil {
		return err
	}
	return ctx.Err()
}

const getSensorsQuery = `select * from db.public.sensors;`

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	rows, err := r.pool.Query(ctx, getSensorsQuery)
	if err != nil {
		return nil, fmt.Errorf("can't select sensors %w", err)
	}
	defer rows.Close()

	sensors := make([]domain.Sensor, 0)
	for rows.Next() {
		sensor := domain.Sensor{}
		if err := scanSensor(&sensor, rows); err != nil {
			return nil, fmt.Errorf("can't scan sensor: %w", err)
		}

		sensors = append(sensors, sensor)
	}

	return sensors, ctx.Err()
}

func scanSensor(sensor *domain.Sensor, row pgx.Row) error {
	return row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity)
}

const getSensorByIDQuery = `select * from db.public.sensors where id=$1`

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, getSensorByIDQuery, id)
	return getSensor(ctx, row)
}

const getSensorBySerialNumberQuery = `select * from db.public.sensors where serial_number=$1`

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, getSensorBySerialNumberQuery, sn)
	return getSensor(ctx, row)
}

func getSensor(ctx context.Context, row pgx.Row) (*domain.Sensor, error) {
	sensor := &domain.Sensor{}
	if err := scanSensor(sensor, row); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, fmt.Errorf("can't scan sensor: %w", err)
	}

	return sensor, ctx.Err()
}
