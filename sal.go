package sal

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sync"

	"github.com/pkg/errors"
)

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

type KeysIntf map[string]interface{}

type RowMap map[string]interface{}

func ProcessQueryAndArgs(query string, reqMap RowMap) (string, []interface{}) {
	pgQuery, argsNames := QueryArgs(query)
	var args = make([]interface{}, 0, len(argsNames))
	for _, name := range argsNames {
		args = append(args, reqMap[name])
	}
	return pgQuery, args
}

type ProcessRower interface {
	ProcessRow(rowMap RowMap)
}

type QueryHandler interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type TransactionBegin interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Txer interface {
	Tx() *WrappedTx
}

type WrappedTx struct {
	Tx   *sql.Tx
	ctrl *Controller
}

func NewWrappedTx(tx *sql.Tx, ctrl *Controller) *WrappedTx {
	return &WrappedTx{Tx: tx, ctrl: ctrl}
}

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

type Controller struct {
	BeforeQuery []BeforeQueryFunc
	sync.RWMutex
	CacheStmts map[string]*sql.Stmt
}

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

func (ctrl *Controller) PrepareStmt(ctx context.Context, qh QueryHandler, query string) (*sql.Stmt, error) {
	var (
		err  error
		stmt *sql.Stmt
	)

	stmt = ctrl.findStmt(query)
	if stmt == nil {
		stmt, err = ctrl.prepareStmt(ctx, qh, query)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to prepare stmt on query %q", query)
		}
		ctrl.putStmt(query, stmt)
	}

	txOpened, _ := ctx.Value(ContextKeyTxOpened).(bool)
	if txOpened {
		txh, ok := qh.(*sql.Tx)
		if !ok {
			return nil, errors.New("failed to get transaction handler")
		}
		stmt = txh.StmtContext(ctx, stmt)
	}

	return stmt, nil
}

type contextKey int

const (
	ContextKeyTxOpened contextKey = iota
	ContextKeyOperationType
	ContextKeyMethodName
)

type ClientOption func(ctrl *Controller)

func BeforeQuery(before ...BeforeQueryFunc) ClientOption {
	return func(ctrl *Controller) { ctrl.BeforeQuery = append(ctrl.BeforeQuery, before...) }
}

type BeforeQueryFunc func(ctx context.Context, query string, req interface{}) (context.Context, FinalizerFunc)

type FinalizerFunc func(ctx context.Context, err error)

type OperationType int

const (
	OperationTypeQueryRow OperationType = iota
	OperationTypeQuery
	OperationTypeExec
	OperationTypeBegin
	OperationTypeCommit
	OperationTypeRollback
	OperationTypePrepare
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

func (op OperationType) String() string {
	return operationTypeNames[op]
}
