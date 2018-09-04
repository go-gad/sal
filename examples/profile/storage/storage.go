package storage

import (
	"context"
	"time"
)

//go:generate salgen -destination=./client.go -package=github.com/go-gad/sal/examples/profile/storage github.com/go-gad/sal/examples/profile/storage Store
type Store interface {
	CreateUser(ctx context.Context, req CreateUserReq) (CreateUserResp, error)
	AllUsers(ctx context.Context, req AllUsersReq) ([]*AllUsersResp, error)
}

type CreateUserReq struct {
	Name  string `sql:"name"`
	Email string `sql:"email"`
}

func (r CreateUserReq) Query() string {
	return `INSERT INTO users(name, email, created_at) VALUES(@name, @email, now()) RETURNING id, created_at`
}

type CreateUserResp struct {
	ID        int64     `sql:"id"`
	CreatedAt time.Time `sql:"created_at"`
}

type AllUsersReq struct{}

func (AllUsersReq) Query() string {
	return `SELECT id, name, email, created_at FROM users`
}

type AllUsersResp struct {
	ID        int64     `sql:"id"`
	Name      string    `sql:"name"`
	Email     string    `sql:"email"`
	CreatedAt time.Time `sql:"created_at"`
}
