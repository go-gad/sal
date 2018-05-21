// Example of package with datastore definitions
package bookstore

import (
	"context"
	"time"
)

type CreateAuthorReq struct {
	Name string
	Desc string
}

func (cr *CreateAuthorReq) Query() string {
	return `INSERT INTO authors (name, desc, created_at) VALUES($1, $2, now()) RETURNING id, created_at`
}

type CreateAuthorResp struct {
	Id        int64
	CreatedAt time.Time
}

type GetAuthorsReq int64

func (r *GetAuthorsReq) Query() string {
	return `SELECT id, created_at, name, desc FROM authors WHERE id>$1`
}

type GetAuthorsResp struct {
	Id        int64
	CreatedAt time.Time
	Name      string
	Desc      string
}

type UpdateAuthorReq struct {
	Id   int64
	Name string
	Desc string
}

type StoreClient interface {
	CreateAuthor(context.Context, *CreateAuthorReq) (*CreateAuthorResp, error)
	GetAuthors(context.Context, GetAuthorsReq) ([]*GetAuthorsResp, error)
	UpdateAuthor(context.Context, *UpdateAuthorReq) error
}
