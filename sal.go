package sal

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sync"

	"github.com/pkg/errors"
)

// reQueryArgs represents the regexp to define named args in query.
var reQueryArgs = regexp.MustCompile("@[A-Za-z0-9_]+")

// QueryArgs receives the query with named arguments
// and returns a query with posgtresql placeholders and a ordered slice named args.
//
// Naive implementation.
func QueryArgs(query string) (string, []string) {
	var args = make([]string, 0)
	pgQuery := reQueryArgs.ReplaceAllStringFunc(query, func(arg string) string {
		args = append(args, arg[1:])
		return fmt.Sprintf("$%d", len(args))
	})
	return pgQuery, args
}

// RowMap contains mapping between column name in database and interface of value.
type RowMap map[string]interface{}

// ProcessQueryAndArgs process query with named args to driver specific query.
func ProcessQueryAndArgs(query string, reqMap RowMap) (string, []interface{}) {
	pgQuery, argsNames := QueryArgs(query)
	var args = make([]interface{}, 0, len(argsNames))
	for _, name := range argsNames {
		args = append(args, reqMap[name])
	}
	return pgQuery, args
}

// ProcessRower is an interface that defines the signature of method of request or response
// that can allow to write pre processor of RowMap values.
//		type GetAuthorsReq struct {
//			ID   int64   `sql:"id"`
//			Tags []int64 `sql:"tags"`
//		}
//
//		func (r GetAuthorsReq) ProcessRow(rowMap sal.RowMap) {
//			rowMap["tags"] = pq.Array(r.Tags)
//		}
// As an argument method receives the RowMap object.
type ProcessRower interface {
	ProcessRow(rowMap RowMap)
}

// QueryHandler describes the methods that are required to pass to constructor of the object
// implementation of user interface.
type QueryHandler interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// TransactionBegin describes the signature of method of user interface to start transaction.
type TransactionBegin interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Txer describes the method to return implementation of Transaction interface.
type Txer interface {
	Tx() *WrappedTx
}

// Transaction is an interface that describes the method to work with transaction object.
// Signature is similar to sql.Tx. The difference is Commit and Rollback methods.
// Its methods work with context.
type Transaction interface {
	QueryHandler
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// WrappedTx is a struct that is an implementation of Transaction interface.
type WrappedTx struct {
	Tx   *sql.Tx
	ctrl *Controller
}

// NewWrappedTx returns the WrappedTx object.
func NewWrappedTx(tx *sql.Tx, ctrl *Controller) *WrappedTx {
	return &WrappedTx{Tx: tx, ctrl: ctrl}
}

// QueryContext executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (wtx *WrappedTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypeQuery.String())
	var (
		resp *sql.Rows
		err  error
	)
	for _, fn := range wtx.ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, query, args)
		if fnz != nil {
			defer func() { fnz(ctx, err) }()
		}
	}

	resp, err = wtx.Tx.QueryContext(ctx, query, args...)

	return resp, err
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (wtx *WrappedTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypeExec.String())
	var (
		resp sql.Result
		err  error
	)
	for _, fn := range wtx.ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, query, args)
		if fnz != nil {
			defer func() { fnz(ctx, err) }()
		}
	}

	resp, err = wtx.Tx.ExecContext(ctx, query, args...)

	return resp, err
}

// PrepareContext creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned statement.
// The caller must call the statement's Close method when the statement is no longer needed.
//
//The provided context is used for the preparation of the statement, not for the execution of the statement.
func (wtx *WrappedTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypePrepare.String())
	var (
		resp *sql.Stmt
		err  error
	)
	for _, fn := range wtx.ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, query, nil)
		if fnz != nil {
			defer func() { fnz(ctx, err) }()
		}
	}

	resp, err = wtx.Tx.PrepareContext(ctx, query)

	return resp, err
}

// Stmt returns a transaction-specific prepared statement from an existing statement.
func (wtx *WrappedTx) StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt {
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypeStmt.String())
	for _, fn := range wtx.ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, "", nil)
		if fnz != nil {
			defer func() { fnz(ctx, nil) }()
		}
	}

	return wtx.Tx.StmtContext(ctx, stmt)
}

// Commit commits the transaction.
func (wtx *WrappedTx) Commit(ctx context.Context) error {
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypeCommit.String())
	ctx = context.WithValue(ctx, ContextKeyMethodName, "Commit")
	var err error
	for _, fn := range wtx.ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, "COMMIT", nil)
		if fnz != nil {
			defer func() { fnz(ctx, err) }()
		}
	}

	err = wtx.Tx.Commit()

	return err
}

// Rollback aborts the transaction.
func (wtx *WrappedTx) Rollback(ctx context.Context) error {
	ctx = context.WithValue(ctx, ContextKeyTxOpened, true)
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypeRollback.String())
	ctx = context.WithValue(ctx, ContextKeyMethodName, "Rollback")
	var err error
	for _, fn := range wtx.ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, "ROLLBACK", nil)
		if fnz != nil {
			defer func() { fnz(ctx, err) }()
		}
	}

	err = wtx.Tx.Rollback()

	return err
}

// Controller is a manager of query processing. Contains the stack of middlewares
// and cache of prepared statements.
type Controller struct {
	BeforeQuery []BeforeQueryFunc
	sync.RWMutex
	CacheStmts map[string]*sql.Stmt
}

// NewController retunes a new object of Controller.
func NewController(options ...ClientOption) *Controller {
	ctrl := &Controller{
		BeforeQuery: []BeforeQueryFunc{},
		CacheStmts:  make(map[string]*sql.Stmt),
	}
	for _, option := range options {
		option(ctrl)
	}
	return ctrl
}

func (ctrl *Controller) findStmt(query string) *sql.Stmt {
	ctrl.RLock()
	stmt, ok := ctrl.CacheStmts[query]
	ctrl.RUnlock()
	if ok {
		return stmt
	}

	return nil
}

func (ctrl *Controller) putStmt(query string, stmt *sql.Stmt) {
	ctrl.Lock()
	ctrl.CacheStmts[query] = stmt
	ctrl.Unlock()
}

func (ctrl *Controller) prepareStmt(ctx context.Context, qh QueryHandler, query string) (*sql.Stmt, error) {
	var err error
	ctx = context.WithValue(ctx, ContextKeyOperationType, OperationTypePrepare.String())
	for _, fn := range ctrl.BeforeQuery {
		var fnz FinalizerFunc
		ctx, fnz = fn(ctx, query, nil)
		if fnz != nil {
			defer func() { fnz(ctx, err) }()
		}
	}

	stmt, err := qh.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

// PrepareStmt returns the prepared statements. If stmt is presented in cache then it will be returned.
// if not, stmt will be prepared and put to cache.
func (ctrl *Controller) PrepareStmt(ctx context.Context, qh QueryHandler, query string) (*sql.Stmt, error) {
	var (
		err  error
		stmt *sql.Stmt
	)

	txOpened, _ := ctx.Value(ContextKeyTxOpened).(bool)
	stmt = ctrl.findStmt(query)
	if stmt == nil && !txOpened {
		stmt, err = ctrl.prepareStmt(ctx, qh, query)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to prepare stmt on query %q", query)
		}
		ctrl.putStmt(query, stmt)
	}

	if txOpened {
		txh, ok := qh.(*sql.Tx)
		if !ok {
			return nil, errors.New("failed to get transaction handler")
		}
		if stmt == nil {
			stmt, err = ctrl.prepareStmt(ctx, txh, query)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to prepare stmt with tx on query %q", query)
			}
		} else {
			stmt = txh.StmtContext(ctx, stmt)
		}
	}

	return stmt, nil
}

type contextKey int

const (
	// ContextKeyTxOpened is a key of bool value in context. If true it means that transaction is opened.
	ContextKeyTxOpened contextKey = iota
	// ContextKeyOperationType is a key of value that describes the operation type. See consts OperationType*.
	ContextKeyOperationType
	// ContextKeyMethodName contains the method name from user interface.
	ContextKeyMethodName
)

// ClientOption sets to controller the optional parameters for clients.
type ClientOption func(ctrl *Controller)

// BeforeQuery sets the BeforeQueryFunc that is executed before the query.
func BeforeQuery(before ...BeforeQueryFunc) ClientOption {
	return func(ctrl *Controller) { ctrl.BeforeQuery = append(ctrl.BeforeQuery, before...) }
}

// BeforeQueryFunc is called before the query execution but after the preparing stmts.
// Returns the FinalizerFunc.
type BeforeQueryFunc func(ctx context.Context, query string, req interface{}) (context.Context, FinalizerFunc)

// FinalizerFunc is executed after the query execution.
type FinalizerFunc func(ctx context.Context, err error)

// OperationType is a datatype for operation types.
type OperationType int

const (
	// OperationTypeQueryRow is a handler.Query operation and single row in response (like db.QueryRow).
	OperationTypeQueryRow OperationType = iota
	// OperationTypeQuery is a handler.Query operation.
	OperationTypeQuery
	// OperationTypeExec is a handler.Exec operation.
	OperationTypeExec
	// OperationTypeBegin is a start transaction operation, db.Begin().
	OperationTypeBegin
	// OperationTypeCommit is a commits the transaction operation, tx.Commit().
	OperationTypeCommit
	// OperationTypeRollback is a aborting the transaction operation, tx.Rollback().
	OperationTypeRollback
	// OperationTypePrepare is a prepare statements operation.
	OperationTypePrepare
	// OperationTypeStmt is a operation of prepare statements on transaction.
	OperationTypeStmt
)

var operationTypeNames = []string{
	"QueryRow",
	"Query",
	"Exec",
	"Begin",
	"Commit",
	"Rollback",
	"Prepare",
	"Stmt",
}

// String returns the string name of operation.
func (op OperationType) String() string {
	return operationTypeNames[op]
}
