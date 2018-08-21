package looker_test

import (
	"reflect"
	"testing"

	pkg_ "github.com/go-gad/sal/examples/bookstore1"
	"github.com/go-gad/sal/looker"
	"github.com/go-gad/sal/looker/testdata"
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
	var typ reflect.Type = reflect.TypeOf((*pkg_.StoreClient)(nil)).Elem()
	intf := looker.LookAtInterface(typ)
	t.Logf("Interface %# v", pretty.Formatter(intf))
}

func TestLookAtParameter(t *testing.T) {
	var typ reflect.Type = reflect.TypeOf(testdata.Req1{})

	if typ.Kind() == reflect.Ptr {
		t.Log("It is a pointer")
		typ = typ.Elem()
	}
	se := looker.LookAtParameter(typ)
	t.Logf("struct element %# v", pretty.Formatter(se))
}

func TestLookAtParameter2(t *testing.T) {
	tf := reflect.TypeOf(testdata.Foo)

	for _, tc := range []struct {
		typ reflect.Type
	}{
		{reflect.TypeOf(testdata.Req1{})},
		{reflect.TypeOf(testdata.List1{})},
		{reflect.TypeOf([]*testdata.Req1{})},
		{tf.In(0)},
	} {
		t.Logf("––––")
		t.Logf("kind[base type] %q", tc.typ.Kind().String())
		t.Logf("string %q", tc.typ.String())
		t.Logf("name %q", tc.typ.Name())
		t.Logf("pkgpath %q", tc.typ.PkgPath())
		if tc.typ.Kind() == reflect.Slice {
			t.Log(">>>")
			el := tc.typ.Elem()
			if el.Kind() == reflect.Ptr {
				el = el.Elem()
			}
			t.Logf("\tkind[base type] %q", el.Kind().String())
			t.Logf("\tstring %q", el.String())
			t.Logf("\tname %q", el.Name())
			t.Logf("\tpkgpath %q", el.PkgPath())
		}

	}

}

func TestLookAtField(t *testing.T) {
	req := testdata.Req1{ID: 4123, Name: "zooloo"}
	var typ reflect.Type = reflect.TypeOf(req)
	ft := typ.Field(0)
	f := looker.LookAtField(ft)
	t.Logf("struct field %# v", pretty.Formatter(f))
}
