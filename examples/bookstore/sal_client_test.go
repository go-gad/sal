package bookstore

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/go-gad/sal"
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
	b1 := func(ctx context.Context, query string, req interface{}) (context.Context, sal.FinalizerFunc) {
		start := time.Now()
		return ctx, func(ctx context.Context, err error) {
			t.Logf(
				"%q > Opeartion %q: %q with req %#v took [%v] inTx[%v] Error: %+v",
				ctx.Value(sal.ContextKeyMethodName),
				ctx.Value(sal.ContextKeyOperationType),
				query,
				req,
				time.Since(start),
				ctx.Value(sal.ContextKeyTxOpened),
				err,
			)
		}
	}

	client := NewStore(db, sal.BeforeQuery(b1))

	req := CreateAuthorReq{BaseAuthor{Name: "foo", Desc: "Bar"}}

	expResp := CreateAuthorResp{ID: 1, CreatedAt: time.Now().Truncate(time.Millisecond)}
	rows := sqlmock.NewRows([]string{"ID", "CreatedAt"}).AddRow(expResp.ID, expResp.CreatedAt)
	mock.ExpectPrepare(`INSERT INTO authors .+`)
	mock.ExpectQuery(`INSERT INTO authors .+`).WithArgs(req.Name, req.Desc).WillReturnRows(rows)

	resp, err := client.CreateAuthor(context.Background(), req)
	assert.Equal(t, expResp, resp)
	assert.Nil(t, mock.ExpectationsWereMet())
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
	client := NewStore(db)

	req := GetAuthorsReq{ID: 123, Tags: Tags{Tags: []int64{33, 44, 55}}}

	expResp := []*GetAuthorsResp{
		&GetAuthorsResp{ID: 10, Name: "Bob", Desc: "d1", Tags: Tags{Tags: []int64{1, 2, 3}}, CreatedAt: time.Now().Truncate(time.Millisecond)},
		&GetAuthorsResp{ID: 20, Name: "Jhn", Desc: "d2", Tags: Tags{Tags: []int64{4, 5, 6}}, CreatedAt: time.Now().Truncate(time.Millisecond)},
		&GetAuthorsResp{ID: 30, Name: "Max", Desc: "d3", Tags: Tags{Tags: []int64{6, 7, 8}}, CreatedAt: time.Now().Truncate(time.Millisecond)},
	}

	rows := sqlmock.NewRows([]string{"id", "created_at", "name", "desc", "tags"})
	for _, v := range expResp {
		rows = rows.AddRow(v.ID, v.CreatedAt, v.Name, v.Desc, dv(v.Tags.Tags))
	}

	mock.ExpectPrepare(`SELECT id, created_at, name,.+`)
	mock.ExpectQuery(`SELECT id, created_at, name,.+`).WithArgs(req.ID, pq.Array(req.Tags.Tags)).WillReturnRows(rows)

	resp, err := client.GetAuthors(context.Background(), req)
	assert.Equal(t, expResp, resp)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSalStore_SameName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := NewStore(db)

	req := SameNameReq{}

	expResp := SameNameResp{
		Bar: "val level 1",
		Foo: Foo{
			Bar: "val level 2",
		},
	}
	var rows *sqlmock.Rows
	{
		rows = sqlmock.NewRows([]string{"Bar", "Bar"})
		rows = rows.AddRow(expResp.Bar, expResp.Foo.Bar)
	}

	mock.ExpectPrepare(`SELECT.+`)
	mock.ExpectQuery(`SELECT.+`).WithArgs().WillReturnRows(rows)

	resp, err := client.SameName(context.Background(), req)
	assert.Equal(t, &expResp, resp)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestSalStore_UpdateAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := NewStore(db)

	req := UpdateAuthorReq{ID: 123, BaseAuthor: BaseAuthor{Name: "John", Desc: "foo-bar"}}

	mock.ExpectPrepare("UPDATE authors SET.+")
	mock.ExpectExec("UPDATE authors SET.+").WithArgs(req.Name, req.Desc, req.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err = client.UpdateAuthor(context.Background(), &req)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestNewStoreController(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	b1 := func(ctx context.Context, query string, req interface{}) (context.Context, sal.FinalizerFunc) {
		start := time.Now()
		return ctx, func(ctx context.Context, err error) {
			t.Logf(
				"%q > Opeartion %q: %q with req %#v took [%v] inTx[%v] Error: %+v",
				ctx.Value(sal.ContextKeyMethodName),
				ctx.Value(sal.ContextKeyOperationType),
				query,
				req,
				time.Since(start),
				ctx.Value(sal.ContextKeyTxOpened),
				err,
			)
		}
	}
	client := NewStore(db, sal.BeforeQuery(b1))

	req1 := CreateAuthorReq{BaseAuthor{Name: "foo", Desc: "Bar"}}
	rows := sqlmock.NewRows([]string{"ID", "CreatedAt"}).AddRow(int64(1), time.Now().Truncate(time.Millisecond))

	req2 := UpdateAuthorReq{ID: 123, BaseAuthor: BaseAuthor{Name: "John", Desc: "foo-bar"}}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO authors .+`) // on connection
	mock.ExpectPrepare(`INSERT INTO authors .+`) // on transaction
	mock.ExpectQuery(`INSERT INTO authors .+`).WithArgs(req1.Name, req1.Desc).WillReturnRows(rows)
	mock.ExpectPrepare("UPDATE authors SET.+") // on connection
	mock.ExpectPrepare("UPDATE authors SET.+") // on transaction
	mock.ExpectExec("UPDATE authors SET.+").WithArgs(req2.Name, req2.Desc, req2.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()

	tx, err := client.BeginTx(ctx, nil)
	assert.Nil(t, err)

	_, err = tx.CreateAuthor(ctx, req1)
	assert.Nil(t, err)

	err = tx.UpdateAuthor(ctx, &req2)
	assert.Nil(t, err)

	err = tx.Tx().Commit(ctx)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())

}

func TestSalStore_GetBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client := NewStore(db)

	req := GetBooksReq{}
	// time.Now().Truncate(time.Millisecond)
	expResp := []*GetBooksResp{
		&GetBooksResp{ID: 10, Title: "foo-10"},
		&GetBooksResp{ID: 20, Title: "foo-20"},
		&GetBooksResp{ID: 30, Title: "foo-30"},
	}
	//Scan(&id, nil, &title, nil)
	rows := sqlmock.NewRows([]string{"id", "created_at", "title", "desc"})
	for _, v := range expResp {
		rows = rows.AddRow(v.ID, time.Now().Truncate(time.Millisecond), v.Title, "trash")
	}

	mock.ExpectPrepare(`SELECT \* FROM books`)
	mock.ExpectQuery(`SELECT \* FROM books`).WillReturnRows(rows)

	resp, err := client.GetBooks(context.Background(), req)
	assert.Equal(t, expResp, resp)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}
