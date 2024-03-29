package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrNilUserPointer = errors.New("nil user is provided")
)

type UserID int64

type UserRepository struct {
	storage map[UserID]*domain.User
	m       sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{storage: map[UserID]*domain.User{}, m: sync.RWMutex{}}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return ErrNilUserPointer
	}
	r.m.Lock()
	r.storage[UserID(user.ID)] = user
	r.m.Unlock()
	return ctx.Err()
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	r.m.RLock()
	user, has := r.storage[UserID(id)]
	r.m.RUnlock()
	if !has {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrUserNotFound
	}
	return user, ctx.Err()
}
