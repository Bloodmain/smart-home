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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

const saveEventQuery = `insert into db.public.users (NAME) values ($1);`

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx, saveEventQuery, user.Name)
	if err != nil {
		return err
	}
	return ctx.Err()
}

const getUserByIDQuery = `select id, name from db.public.users where id=$1`

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, getUserByIDQuery, id)

	user := &domain.User{}
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, fmt.Errorf("can't scan user: %w", err)
	}

	return user, ctx.Err()
}
