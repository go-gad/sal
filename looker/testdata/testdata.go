package testdata

import (
	"context"

	"github.com/go-gad/sal"
)

type Req1 struct {
	ID   int64 `sql:"id"`
	Name string
}

type List1 []*Req1

func Foo(ctx context.Context, req []*Req1) error { return nil }

type Req2 struct {
	ID int64 `sql:"id"`
}

func (r *Req2) ProcessRow(rm sal.RowMap) {}
