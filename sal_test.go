package sal

import (
	"context"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestQueryArgs(t *testing.T) {
	//t.Skip("todo")
	var tt = []struct {
		QueryNamed string
		QueryPg    string
		NamedArgs  []string
	}{
		{
			QueryNamed: `UPDATE authors SET name=@name, desc=@desc WHERE id=@id`,
			QueryPg:    `UPDATE authors SET name=$1, desc=$2 WHERE id=$3`,
			NamedArgs:  []string{"name", "desc", "id"},
		}, {
			QueryNamed: `SELECT id, created_at, name, desc FROM authors WHERE id>@id`,
			QueryPg:    `SELECT id, created_at, name, desc FROM authors WHERE id>$1`,
			NamedArgs:  []string{"id"},
		},
	}
	for _, tc := range tt {
		query, args := QueryArgs(tc.QueryNamed)
		assert.Equal(t, tc.QueryPg, query)
		assert.Equal(t, tc.NamedArgs, args)
	}
}

func TestMapIndex_NextVal(t *testing.T) {
	ind := make(mapIndex)
	assert.Equal(t, 0, ind.NextVal("foo"))
	assert.Equal(t, 1, ind.NextVal("foo"))
	assert.Equal(t, 2, ind.NextVal("foo"))
	assert.Equal(t, 0, ind.NextVal("bar"))
	assert.Equal(t, 3, ind.NextVal("foo"))
	assert.Equal(t, 1, ind.NextVal("bar"))
}

func TestRowMap(t *testing.T) {
	assert := assert.New(t)
	rm := make(RowMap)
	assert.Nil(rm.Get("foo"))
	assert.Nil(rm.GetByIndex("foo", 0))
	rm.AppendTo("foo", 777)
	assert.Equal(777, rm.Get("foo"))
	assert.Equal(777, rm.GetByIndex("foo", 0))
	assert.Nil(rm.GetByIndex("foo", 1))
}

func TestGetDests(t *testing.T) {
	assert := assert.New(t)
	rm := make(RowMap)
	rm.AppendTo("id", 111)
	rm.AppendTo("title", "foobar")
	cols := []string{"id", "created_at", "title", "desc"}
	dest := GetDests(cols, rm)
	var n skippedField
	expt := []interface{}{111, &n, "foobar", &n}
	assert.Equal(expt, dest)
}

func TestController_PrepareStmt(t *testing.T) {
	assert := assert.New(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	ctrl := NewController()
	ctx := context.Background()

	// PrepareStmt on dbConn
	query := "SELECT 1"
	mock.ExpectPrepare(query) // on connection
	_, err = ctrl.PrepareStmt(ctx, nil, db, query)
	assert.NoError(err)

	// Checking prepared stmt in cache
	stmt := ctrl.findStmt(query)
	assert.NotNil(stmt)

	// Second try with same query on dbConn returns cached stmt
	stmt2, err := ctrl.PrepareStmt(ctx, nil, db, query)
	assert.NoError(err)
	assert.Equal(stmt, stmt2)

	// If something wrong then returns error
	mock.ExpectPrepare(`SELECT 2`).WillReturnError(errors.New("bye"))
	_, err = ctrl.PrepareStmt(ctx, nil, db, "SELECT 2")
	assert.Error(err)

	// open tx connection
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(err)
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)

	// try to prepare stmt with dbconn in tx context init an error
	_, err = ctrl.PrepareStmt(ctx, nil, db, "SELECT 3")
	assert.Error(err)
	//t.Logf("error %+v", err)

	// non-cached stmt for query, prepare query on tx
	query = "SELECT 4"
	mock.ExpectPrepare(query) // on transaction tx.Stmt
	_, err = ctrl.PrepareStmt(ctx, nil, tx, query)
	assert.NoError(err)

	// something wrong and error
	query = "SELECT 41"
	mock.ExpectPrepare(query).WillReturnError(errors.New("ops")) // on transaction tx.Stmt
	_, err = ctrl.PrepareStmt(ctx, nil, tx, query)
	assert.Error(err)

	// if we call BeginTx and after try to perform prepare query
	// and cache doesn't contain value
	// then stmt is prepared on dbConn and applied to tx
	query = "SELECT 5"
	mock.ExpectPrepare(query) // on connection
	mock.ExpectPrepare(query) // on transaction tx.Stmt
	_, err = ctrl.PrepareStmt(ctx, db, tx, query)
	assert.NoError(err)

	// error case
	query = "SELECT 51"
	mock.ExpectPrepare(query).WillReturnError(errors.New("ops")) // on connection
	//mock.ExpectPrepare(query) // on transaction tx.Stmt
	_, err = ctrl.PrepareStmt(ctx, db, tx, query)
	assert.Error(err)

	// second try with first stmt
	query = "SELECT 1"
	//mock.ExpectPrepare(query) // on transaction tx.Stmt
	_, err = ctrl.PrepareStmt(ctx, db, tx, query)
	assert.NoError(err)

	assert.Nil(mock.ExpectationsWereMet())
}
