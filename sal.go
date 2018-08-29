package sal

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
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

type Transaction interface {
	Tx() TxHandler
}

type TxHandler interface {
	QueryHandler
	TransactionEnd
}

type QueryHandler interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type TransactionBegin interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type TransactionEnd interface {
	Commit() error
	Rollback() error
}

type Controller struct {
	BeforeQuery []BeforeQueryFunc
}

func NewController(options ...ClientOption) *Controller {
	ctrl := &Controller{
		BeforeQuery: []BeforeQueryFunc{},
	}
	for _, option := range options {
		option(ctrl)
	}
	return ctrl
}

type ClientOption func(ctrl *Controller)

func BeforeQuery(before ...BeforeQueryFunc) ClientOption {
	return func(ctrl *Controller) { ctrl.BeforeQuery = append(ctrl.BeforeQuery, before...) }
}

type BeforeQueryFunc func(ctx context.Context, query string, args []interface{}) (context.Context, AfterQueryFunc)

type AfterQueryFunc func(ctx context.Context, err error)
