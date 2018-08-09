package bookstore1

import (
	"context"
	"time"
)

type StoreClient interface {
	CreateAuthor(context.Context, CreateAuthorReq) (*CreateAuthorResp, error)
}

type CreateAuthorReq struct {
	Name string
	Desc string
}

func (cr *CreateAuthorReq) Query() string {
	return `INSERT INTO authors (name, desc, created_at) VALUES(@name, @desc, now()) RETURNING id, created_at`
}

type CreateAuthorResp struct {
	Id        int64
	CreatedAt time.Time
}
