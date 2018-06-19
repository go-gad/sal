// Example of package with datastore definitions. Package contains pretty simple structs.
// There are examples that invoke Query with single and multiple rows, Exec of db connection.
package bookstore

import (
	"context"
	"time"
)

//go:generate salgen -destination=./act/sal_client.go -package=act github.com/go-gad/sal/internal/bookstore StoreClient
type StoreClient interface {
	CreateAuthor(context.Context, CreateAuthorReq) (*CreateAuthorResp, error)
	//GetAuthors(context.Context, GetAuthorsReq) ([]*GetAuthorsResp, error)
	UpdateAuthor(context.Context, *UpdateAuthorReq) error
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

type GetAuthorsReq struct {
	Id int64
}

func (r *GetAuthorsReq) Query() string {
	return `SELECT id, created_at, name, desc FROM authors WHERE id>@id`
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

func (r *UpdateAuthorReq) Query() string {
	return `UPDATE authors SET name=@name, desc=@desc WHERE id=@id`
}
