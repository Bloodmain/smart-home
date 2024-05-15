package postgres

import (
	"context"
	"fmt"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorOwnerRepository struct {
	pool *pgxpool.Pool
}

func NewSensorOwnerRepository(pool *pgxpool.Pool) *SensorOwnerRepository {
	return &SensorOwnerRepository{
		pool,
	}
}

const saveSensorOwnerQuery = `insert into db.public.sensors_users (sensor_id, user_id) values ($1, $2);`

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	_, err := r.pool.Exec(ctx, saveSensorOwnerQuery, sensorOwner.SensorID, sensorOwner.UserID)
	if err != nil {
		return err
	}
	return ctx.Err()
}

const getSensorsByUserId = `select sensor_id, user_id from db.public.sensors_users where user_id = $1;`

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	rows, err := r.pool.Query(ctx, getSensorsByUserId, userID)
	if err != nil {
		return nil, fmt.Errorf("can't select sensors by user id %d %w", userID, err)
	}
	defer rows.Close()

	sensors := make([]domain.SensorOwner, 0)
	for rows.Next() {
		sensor := domain.SensorOwner{}
		if err := rows.Scan(&sensor.SensorID, &sensor.UserID); err != nil {
			return nil, fmt.Errorf("can't scan sensor owner: %w", err)
		}

		sensors = append(sensors, sensor)
	}

	return sensors, ctx.Err()
}
