package bookstore

import (
	"context"
	"time"

	"github.com/go-gad/sal"
	"github.com/lib/pq"
)

//go:generate salgen -destination=./repo/sal_client.go -package=repo github.com/go-gad/sal/examples/bookstore Store
type Store interface {
	Dup() Store
	sal.Controller

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
	ID   int64   `sql:"id"`
	Tags []int64 `sql:"tags"`
}

func (r GetAuthorsReq) ProcessRow(rowMap sal.RowMap) {
	rowMap["tags"] = pq.Array(r.Tags)
}

func (r *GetAuthorsReq) Query() string {
	return `SELECT id, created_at, name, desc, tags FROM authors WHERE id>@id AND tags @> @tags`
}

type GetAuthorsResp struct {
	ID        int64     `sql:"id"`
	CreatedAt time.Time `sql:"created_at"`
	Name      string    `sql:"name"`
	Desc      string    `sql:"desc"`
	Tags      []int64   `sql:"tags"`
}

func (r *GetAuthorsResp) ProcessRow(rowMap sal.RowMap) {
	rowMap["tags"] = pq.Array(&r.Tags)
}

type UpdateAuthorReq struct {
	ID   int64
	Name string
	Desc string
}

func (r *UpdateAuthorReq) Query() string {
	return `UPDATE authors SET Name=@Name, Desc=@Desc WHERE ID=@ID`
}
