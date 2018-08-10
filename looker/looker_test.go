package looker_test

import (
	"reflect"
	"testing"

	pkg_ "github.com/go-gad/sal/examples/bookstore1"
	"github.com/go-gad/sal/looker"
	"github.com/kr/pretty"
)

func TestLookAtInterfaces(t *testing.T) {
	pkgPath := "github.com/go-gad/sal/examples/bookstore1"
	var list = []reflect.Type{
		reflect.TypeOf((*pkg_.StoreClient)(nil)).Elem(),
	}
	pkg := looker.LookAtInterfaces(pkgPath, list)

	t.Logf("package %# v", pretty.Formatter(pkg))
}

func TestLookAtInterface(t *testing.T) {
	pf := getLogger(t)
	var typ reflect.Type = reflect.TypeOf((*pkg_.StoreClient)(nil)).Elem()
	intf := looker.LookAtInterface(typ)
	pf("Interface %# v", pretty.Formatter(intf))
}

func getLogger(t *testing.T) func(string, ...interface{}) {
	return func(s string, kv ...interface{}) {
		t.Logf(s, kv...)
	}
}

type Req1 struct {
	ID   int64
	Name string
}

func TestLookAtParameter(t *testing.T) {
	req := &Req1{ID: 4123, Name: "zooloo"}
	var typ reflect.Type = reflect.TypeOf(req)

	if typ.Kind() == reflect.Ptr {
		t.Log("It is a pointer")
		typ = typ.Elem()
	}

}
