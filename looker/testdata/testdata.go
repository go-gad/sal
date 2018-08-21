package testdata

import "context"

type Req1 struct {
	ID   int64
	Name string
}

type List1 []*Req1

func Foo(ctx context.Context, req []*Req1) error { return nil }
