package looker_test

import (
	"reflect"
	"testing"

	pkg_ "github.com/go-gad/sal/examples/bookstore1"
	"github.com/go-gad/sal/looker"
	"github.com/go-gad/sal/looker/testdata"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
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
	ftyp := reflect.TypeOf(testdata.Foo)

	for _, tc := range []struct {
		test string
		typ  reflect.Type
		prm  looker.Parameter
		kind string
		name string
		ptr  bool
	}{
		{
			test: "user struct",
			typ:  reflect.TypeOf(testdata.Req1{}),
			prm: &looker.StructElement{
				ImportPath: looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata"},
				UserType:   "Req1",
				IsPointer:  false,
			},
			kind: reflect.Struct.String(),
			name: "testdata.Req1",
			ptr:  false,
		}, {
			test: "slice of user structs",
			typ:  reflect.TypeOf([]*testdata.Req1{}),
			prm: &looker.StructElement{
				ImportPath: looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata"},
				UserType:   "Req1",
				IsPointer:  false,
			},
			kind: reflect.Slice.String(),
			name: "[]*testdata.Req1",
			ptr:  false,
		}, {
			test: "user type of slice",
			typ:  reflect.TypeOf(testdata.List1{}),
			prm: &looker.StructElement{
				ImportPath: looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata"},
				UserType:   "List1",
				IsPointer:  false,
			},
			kind: reflect.Slice.String(),
			name: "testdata.List1",
			ptr:  false,
		}, {
			test: "context",
			typ:  ftyp.In(0),
			prm: &looker.InterfaceElement{
				ImportPath: looker.ImportElement{Path: "context"},
				UserType:   "Context",
			},
			kind: reflect.Interface.String(),
			name: "context.Context",
			ptr:  false,
		}, {
			test: "error",
			typ:  ftyp.Out(0),
			prm: &looker.InterfaceElement{
				ImportPath: looker.ImportElement{Path: ""},
				UserType:   "error",
			},
			kind: reflect.Interface.String(),
			name: "error",
			ptr:  false,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			assert := assert.New(t)
			prm := looker.LookAtParameter(tc.typ)
			assert.Equal(tc.kind, prm.Kind())
			assert.Equal(tc.name, prm.Name())
			assert.Equal(tc.ptr, prm.Pointer())
		})
	}
}

func TestLookAtFields(t *testing.T) {
	var typ reflect.Type = reflect.TypeOf(testdata.Req1{})
	actFields := looker.LookAtFields(typ)

	expFields := looker.Fields{
		{
			Name:       "ID",
			ImportPath: looker.ImportElement{},
			BaseType:   "int64",
			UserType:   "int64",
			Anonymous:  false,
			Tag:        "id",
		},
		{
			Name:       "Name",
			ImportPath: looker.ImportElement{},
			BaseType:   "string",
			UserType:   "string",
			Anonymous:  false,
			Tag:        "",
		},
	}
	assert.Equal(t, expFields, actFields)

	t.Logf("struct field %# v", pretty.Formatter(actFields))
}
