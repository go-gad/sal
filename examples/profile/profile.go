package profile

import (
	"context"
	"time"

	"github.com/go-gad/sal/examples/profile/storage"
)

type UserService struct {
	store storage.Store
}

func NewUserService(store storage.Store) *UserService {
	return &UserService{store: store}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*User, error) {
	req := storage.CreateUserReq{
		Name:  name,
		Email: email,
	}

	resp, err := s.store.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        resp.ID,
		Name:      name,
		Email:     email,
		CreatedAt: resp.CreatedAt,
	}, nil
}

func (s *UserService) AllUsers(ctx context.Context) (Users, error) {
	resp, err := s.store.AllUsers(ctx, storage.AllUsersReq{})
	if err != nil {
		return nil, err
	}
	users := make(Users, 0, len(resp))
	for _, v := range resp {
		users = append(users, &User{
			ID:        v.ID,
			Name:      v.Name,
			Email:     v.Email,
			CreatedAt: v.CreatedAt,
		})
	}

	return users, nil
}

type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
}

type Users []*User
