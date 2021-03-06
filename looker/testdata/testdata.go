package testdata

import (
	"context"

	"github.com/go-gad/sal"
	"github.com/go-gad/sal/looker/testdata/foo-bar"
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

type Lvl1 struct {
	Name string
	Desc string
	Lvl21
	Lvl22
}

type Lvl21 struct {
	Foo string
	Bar string
}
type Lvl22 struct {
	Foo string
	Bar string
	Lvl3
}

type Lvl3 struct {
	Foo string
	Bar string
}

type Store interface {
	UpdateAuthor(context.Context, *foo.Body) error
}
