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

type UserRepository struct {
	storage []*domain.User
	m       sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{storage: []*domain.User{}, m: sync.RWMutex{}}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return ErrNilUserPointer
	}
	r.m.Lock()
	r.storage = append(r.storage, user)
	r.m.Unlock()
	return ctx.Err()
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	r.m.RLock()
	for _, v := range r.storage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if v.ID == id {
				return v, ctx.Err()
			}
		}
	}
	r.m.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, ErrUserNotFound
	}
}
