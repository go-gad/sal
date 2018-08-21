package actsal

import (
	"context"
	"testing"
	"time"

	"github.com/go-gad/sal/examples/bookstore1"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestSalStoreClient_CreateAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := NewStoreClient(db)

	req := bookstore1.CreateAuthorReq{Name: "foo", Desc: "Bar"}

	expResp := bookstore1.CreateAuthorResp{ID: 1, CreatedAt: time.Now().Truncate(time.Millisecond)}
	rows := sqlmock.NewRows([]string{"ID", "CreatedAt"}).AddRow(expResp.ID, expResp.CreatedAt)
	mock.ExpectQuery(`INSERT INTO authors .+`).WithArgs(req.Name, req.Desc).WillReturnRows(rows)

	resp, err := client.CreateAuthor(context.Background(), req)
	assert.Equal(t, &expResp, resp)
}

func TestSalStoreClient_GetAuthors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := NewStoreClient(db)

	req := bookstore1.GetAuthorsReq{ID: 123, Tags: []int64{33, 44, 55}}

	expResp := []*bookstore1.GetAuthorsResp{
		&bookstore1.GetAuthorsResp{ID: 10, Name: "Bob", Desc: "d1", CreatedAt: time.Now().Truncate(time.Millisecond)},
		&bookstore1.GetAuthorsResp{ID: 20, Name: "Jhn", Desc: "d2", CreatedAt: time.Now().Truncate(time.Millisecond)},
		&bookstore1.GetAuthorsResp{ID: 30, Name: "Max", Desc: "d3", CreatedAt: time.Now().Truncate(time.Millisecond)},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "desc", "created_at"})
	for _, v := range expResp {
		rows = rows.AddRow(v.ID, v.Name, v.Desc, v.CreatedAt)
	}

	mock.ExpectQuery(`SELECT id, created_at, name,.+`).WithArgs(req.ID, pq.Array(req.Tags)).WillReturnRows(rows)

	resp, err := client.GetAuthors(context.Background(), req)
	assert.Equal(t, expResp, resp)
	assert.Nil(t, err)
}

func TestSalStoreClient_UpdateAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := NewStoreClient(db)

	req := bookstore1.UpdateAuthorReq{ID: 123, Name: "John", Desc: "foo-bar"}
	mock.ExpectExec("UPDATE authors SET.+").WithArgs(req.Name, req.Desc, req.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err = client.UpdateAuthor(context.Background(), &req)
	assert.Nil(t, err)
}
