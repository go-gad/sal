package looker_test

import (
	"reflect"
	"testing"

	pkg_ "github.com/go-gad/sal/examples/bookstore1"
	"github.com/go-gad/sal/looker"
)

func TestLookAtInterfaces(t *testing.T) {
	pf := getLogger(t)
	pkgPath := "github.com/go-gad/sal/examples/bookstore1"
	var list = []reflect.Type{
		reflect.TypeOf((*pkg_.StoreClient)(nil)).Elem(),
	}
	pkg := looker.LookAtInterfaces(pkgPath, list)
	pf("package %#v", pkg)
}

func TestLookAtInterface(t *testing.T) {
	pf := getLogger(t)
	var typ reflect.Type = reflect.TypeOf((*pkg_.StoreClient)(nil)).Elem()
	intf := looker.LookAtInterface(typ)
	pf("Interface %#v", intf)
	for i, v := range intf.Methods {
		pf("\t[%d] method %q", i, v.Name)
		for _, prm := range v.In {
			pf("\t\tparam IN %#v", prm)
			if prm.Type() == looker.ParameterTypeStruct {
				sprm := prm.(*looker.ParameterStruct)
				for _, f := range sprm.Fields {
					pf("\t\t\tfield %#v", f)
				}
			}
		}
		for _, prm := range v.Out {
			pf("\t\tparam OUT %#v", prm)
			if prm.Type() == looker.ParameterTypeStruct {
				sprm := prm.(*looker.ParameterStruct)
				for _, f := range sprm.Fields {
					pf("\t\t\tfield %#v", f)
				}
			}
		}
	}

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
