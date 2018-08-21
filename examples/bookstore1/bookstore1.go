package bookstore1

import (
	"context"
	"time"
)

//go:generate salgen -destination=./actsal/sal_client.go -package=actsal github.com/go-gad/sal/examples/bookstore1 StoreClient
type StoreClient interface {
	CreateAuthor(context.Context, CreateAuthorReq) (*CreateAuthorResp, error)
	GetAuthors(context.Context, GetAuthorsReq) ([]*GetAuthorsResp, error)
	UpdateAuthor(context.Context, *UpdateAuthorReq) error
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

type GetAuthorsReq struct {
	ID int64 `sql:"id"`
}

func (r *GetAuthorsReq) Query() string {
	return `SELECT id, created_at, name, desc FROM authors WHERE id>@id`
}

type GetAuthorsResp struct {
	ID        int64     `sql:"id"`
	CreatedAt time.Time `sql:"created_at"`
	Name      string    `sql:"name"`
	Desc      string    `sql:"desc"`
}

type UpdateAuthorReq struct {
	ID   int64
	Name string
	Desc string
}

func (r *UpdateAuthorReq) Query() string {
	return `UPDATE authors SET Name=@Name, Desc=@Desc WHERE ID=@ID`
}
