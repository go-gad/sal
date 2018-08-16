package actsal

import (
	"context"
	"testing"
	"time"

	"github.com/go-gad/sal/examples/bookstore1"
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
