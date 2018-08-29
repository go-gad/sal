package repo_test

import (
	"context"
	"testing"
	"time"

	"database/sql/driver"

	"github.com/go-gad/sal"
	"github.com/go-gad/sal/examples/bookstore"
	"github.com/go-gad/sal/examples/bookstore/repo"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestSalStore_CreateAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b1 := func(ctx context.Context, query string, args []interface{}) (context.Context, sal.AfterQueryFunc) {
		t.Logf("Query %q with args %#v", query, args)
		return ctx, func(ctx context.Context, err error) {
			t.Logf("Error: %+v", err)
		}
	}

	client := repo.NewStore(db, sal.BeforeQuery(b1))

	req := bookstore.CreateAuthorReq{Name: "foo", Desc: "Bar"}

	expResp := bookstore.CreateAuthorResp{ID: 1, CreatedAt: time.Now().Truncate(time.Millisecond)}
	rows := sqlmock.NewRows([]string{"ID", "CreatedAt"}).AddRow(expResp.ID, expResp.CreatedAt)
	mock.ExpectQuery(`INSERT INTO authors .+`).WithArgs(req.Name, req.Desc).WillReturnRows(rows)

	resp, err := client.CreateAuthor(context.Background(), req)
	assert.Equal(t, &expResp, resp)
}

func dv(a []int64) driver.Value {
	v, _ := pq.Int64Array(a).Value()
	return v
}

func TestSalStore_GetAuthors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := repo.NewStore(db)

	req := bookstore.GetAuthorsReq{ID: 123, Tags: []int64{33, 44, 55}}

	expResp := []*bookstore.GetAuthorsResp{
		&bookstore.GetAuthorsResp{ID: 10, Name: "Bob", Desc: "d1", Tags: []int64{1, 2, 3}, CreatedAt: time.Now().Truncate(time.Millisecond)},
		&bookstore.GetAuthorsResp{ID: 20, Name: "Jhn", Desc: "d2", Tags: []int64{4, 5, 6}, CreatedAt: time.Now().Truncate(time.Millisecond)},
		&bookstore.GetAuthorsResp{ID: 30, Name: "Max", Desc: "d3", Tags: []int64{6, 7, 8}, CreatedAt: time.Now().Truncate(time.Millisecond)},
	}

	rows := sqlmock.NewRows([]string{"id", "created_at", "name", "desc", "tags"})
	for _, v := range expResp {
		rows = rows.AddRow(v.ID, v.CreatedAt, v.Name, v.Desc, dv(v.Tags))
	}

	mock.ExpectQuery(`SELECT id, created_at, name,.+`).WithArgs(req.ID, pq.Array(req.Tags)).WillReturnRows(rows)

	resp, err := client.GetAuthors(context.Background(), req)
	assert.Equal(t, expResp, resp)
	assert.Nil(t, err)
}

func TestSalStore_UpdateAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := repo.NewStore(db)

	req := bookstore.UpdateAuthorReq{ID: 123, Name: "John", Desc: "foo-bar"}
	mock.ExpectExec("UPDATE authors SET.+").WithArgs(req.Name, req.Desc, req.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err = client.UpdateAuthor(context.Background(), &req)
	assert.Nil(t, err)
}

func TestNewStoreController(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := repo.NewStore(db)

	req1 := bookstore.CreateAuthorReq{Name: "foo", Desc: "Bar"}
	rows := sqlmock.NewRows([]string{"ID", "CreatedAt"}).AddRow(int64(1), time.Now().Truncate(time.Millisecond))

	req2 := bookstore.UpdateAuthorReq{ID: 123, Name: "John", Desc: "foo-bar"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO authors .+`).WithArgs(req1.Name, req1.Desc).WillReturnRows(rows)
	mock.ExpectExec("UPDATE authors SET.+").WithArgs(req2.Name, req2.Desc, req2.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()

	tx, err := client.BeginTx(ctx, nil)
	assert.Nil(t, err)

	_, err = tx.CreateAuthor(ctx, req1)
	assert.Nil(t, err)

	err = tx.UpdateAuthor(ctx, &req2)
	assert.Nil(t, err)

	err = tx.Tx().Commit()
	assert.Nil(t, err)

}
