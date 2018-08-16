package bookstore1

import (
	"context"
	"time"
)

//go:generate salgen -destination=./actsal/sal_client.go -package=actsal github.com/go-gad/sal/examples/bookstore1 StoreClient
type StoreClient interface {
	CreateAuthor(context.Context, CreateAuthorReq) (*CreateAuthorResp, error)
}

type CreateAuthorReq struct {
	Name string
	Desc string
}

func (cr *CreateAuthorReq) Query() string {
	return `INSERT INTO authors (Name, Desc, CreatedAt) VALUES(@Name, @Desc, now()) RETURNING ID, CreatedAt`
}

type CreateAuthorResp struct {
	ID        int64
	CreatedAt time.Time
}
