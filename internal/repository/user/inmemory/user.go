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
	done := make(chan struct{})
	found := make(chan *domain.User)

	go func() {
		r.m.RLock()
	outer:
		for _, v := range r.storage {
			select {
			case <-ctx.Done():
				break outer
			default:

				if v.ID == id {
					found <- v
					break
				}
			}
		}
		r.m.RUnlock()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s := <-found:
		return s, nil
	case <-done:
		return nil, ErrUserNotFound
	}
}
