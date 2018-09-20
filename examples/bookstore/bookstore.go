package bookstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-gad/sal"
	"github.com/lib/pq"
)

//go:generate salgen -destination=./sal_client.go -package=github.com/go-gad/sal/examples/bookstore github.com/go-gad/sal/examples/bookstore Store
type Store interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Store, error)
	sal.Txer

	CreateAuthor(context.Context, CreateAuthorReq) (CreateAuthorResp, error)
	CreateAuthorPtr(context.Context, CreateAuthorReq) (*CreateAuthorResp, error)
	GetAuthors(context.Context, GetAuthorsReq) ([]*GetAuthorsResp, error)
	UpdateAuthor(context.Context, *UpdateAuthorReq) error
}

type BaseAuthor struct {
	Name string
	Desc string
}

type CreateAuthorReq struct {
	BaseAuthor
}

func (cr *CreateAuthorReq) Query() string {
	return `INSERT INTO authors (Name, Desc, CreatedAt) VALUES(@Name, @Desc, now()) RETURNING ID, CreatedAt`
}

type CreateAuthorResp struct {
	ID        int64
	CreatedAt time.Time
}

type Tags struct {
	Tags []int64 `sql:"tags"`
}

type GetAuthorsReq struct {
	ID int64 `sql:"id"`
	Tags
}

func (r GetAuthorsReq) ProcessRow(rowMap sal.RowMap) {
	rowMap.Set("tags", pq.Array(r.Tags.Tags))
}

func (r *GetAuthorsReq) Query() string {
	return `SELECT id, created_at, name, desc, tags FROM authors WHERE id>@id AND tags @> @tags`
}

type GetAuthorsResp struct {
	ID        int64     `sql:"id"`
	CreatedAt time.Time `sql:"created_at"`
	Name      string    `sql:"name"`
	Desc      string    `sql:"desc"`
	Tags
}

func (r *GetAuthorsResp) ProcessRow(rowMap sal.RowMap) {
	rowMap.Set("tags", pq.Array(&r.Tags.Tags))
}

type UpdateAuthorReq struct {
	ID int64
	BaseAuthor
}

func (r *UpdateAuthorReq) Query() string {
	return `UPDATE authors SET Name=@Name, Desc=@Desc WHERE ID=@ID`
}
