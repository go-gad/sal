package looker_test

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	pkg_ "github.com/go-gad/sal/examples/bookstore"
	"github.com/go-gad/sal/looker"
	"github.com/go-gad/sal/looker/testdata"
	"github.com/go-gad/sal/looker/testdata/foo-bar"
	"github.com/kr/pretty"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
)

var update bool = false

func TestLookAtInterfaces(t *testing.T) {
	pkgPath := "github.com/go-gad/sal/examples/bookstore"
	var list = []reflect.Type{
		reflect.TypeOf((*pkg_.Store)(nil)).Elem(),
	}
	pkg := looker.LookAtInterfaces(pkgPath, list)

	//t.Logf("package %# v", pretty.Formatter(pkg))

	act := fmt.Sprintf("%# v", pretty.Formatter(pkg))
	if update {
		if err := ioutil.WriteFile("testdata/package.golden", []byte(act), 0666); err != nil {
			t.Fatal(err)
		}
	}
	exp, err := ioutil.ReadFile("testdata/package.golden")
	if err != nil {
		t.Fatal(err)
	}
	if string(exp) != act {
		t.Error("actual package is not equal to expected")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(exp), act, true)
		t.Log(dmp.DiffPrettyText(diffs))
	}
}

func TestLookAtInterface(t *testing.T) {
	var typ reflect.Type = reflect.TypeOf((*pkg_.Store)(nil)).Elem()
	intf := looker.LookAtInterface(typ)
	//t.Logf("Interface %# v", pretty.Formatter(intf))
	act := fmt.Sprintf("%# v", pretty.Formatter(intf))
	if update {
		if err := ioutil.WriteFile("testdata/interface.golden", []byte(act), 0666); err != nil {
			t.Fatal(err)
		}
	}

	exp, err := ioutil.ReadFile("testdata/interface.golden")
	if err != nil {
		t.Fatal(err)
	}
	if string(exp) != act {
		t.Error("actual interface is not equal to expected")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(exp), act, true)
		t.Log(dmp.DiffPrettyText(diffs))
	}
}

func TestLookAtParameter(t *testing.T) {
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
			typ:  reflect.TypeOf((*context.Context)(nil)).Elem(),
			prm: &looker.InterfaceElement{
				ImportPath: looker.ImportElement{Path: "context"},
				UserType:   "Context",
			},
			kind: reflect.Interface.String(),
			name: "context.Context",
			ptr:  false,
		}, {
			test: "error",
			typ:  reflect.TypeOf((*error)(nil)).Elem(),
			prm: &looker.InterfaceElement{
				ImportPath: looker.ImportElement{Path: ""},
				UserType:   "error",
			},
			kind: reflect.Interface.String(),
			name: "error",
			ptr:  false,
		}, {
			test: "alias",
			typ:  reflect.TypeOf(foo.Body{}),
			prm: &looker.StructElement{
				ImportPath: looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata/foo-bar", Alias: "foo"},
				UserType:   "Body",
				IsPointer:  false,
			},
			kind: reflect.Struct.String(),
			name: "foo.Body",
			ptr:  false,
		}, {
			test: "alias for slice",
			typ:  reflect.TypeOf(foo.List{}),
			prm: &looker.SliceElement{
				ImportPath: looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata/foo-bar", Alias: "foo"},
				UserType:   "List",
				Item: &looker.StructElement{
					ImportPath:   looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata/foo-bar", Alias: "foo"},
					UserType:     "Body",
					IsPointer:    true,
					ProcessRower: false,
				},
				IsPointer: false,
			},
			kind: reflect.Slice.String(),
			name: "foo.List",
			ptr:  false,
		}, {
			test: "result",
			typ:  reflect.TypeOf((*sql.Result)(nil)).Elem(),
			prm:  &looker.InterfaceElement{},
			kind: reflect.Interface.String(),
			name: "sql.Result",
			ptr:  false,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			dstPkg := looker.ImportElement{Path: "github.com/go-gad/sal/looker"}
			assert := assert.New(t)
			prm := looker.LookAtParameter(tc.typ)
			assert.Equal(tc.kind, prm.Kind())
			assert.Equal(tc.name, prm.Name(dstPkg.Path))
			assert.Equal(tc.ptr, prm.Pointer())
			//t.Logf("element %# v", pretty.Formatter(prm))
		})
	}
}

func TestLookAtFields(t *testing.T) {
	t.Run("common", func(t *testing.T) {
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
				Parents:    []string{},
			},
			{
				Name:       "Name",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{},
			},
		}
		assert.Equal(t, expFields, actFields)
		t.Logf("struct field %# v", pretty.Formatter(actFields))
	})

	t.Run("nested", func(t *testing.T) {
		var typ reflect.Type = reflect.TypeOf(testdata.Lvl1{})
		actFields := looker.LookAtFields(typ)
		expFields := looker.Fields{
			{
				Name:       "Name",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{},
			},
			{
				Name:       "Desc",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{},
			},
			{
				Name:       "Foo",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{"Lvl21"},
			},
			{
				Name:       "Bar",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{"Lvl21"},
			},
			{
				Name:       "Foo",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{"Lvl22"},
			},
			{
				Name:       "Bar",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{"Lvl22"},
			},
			{
				Name:       "Foo",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{"Lvl22", "Lvl3"},
			},
			{
				Name:       "Bar",
				ImportPath: looker.ImportElement{},
				BaseType:   "string",
				UserType:   "string",
				Anonymous:  false,
				Tag:        "",
				Parents:    []string{"Lvl22", "Lvl3"},
			},
		}
		assert.Equal(t, expFields, actFields)
		t.Logf("struct field %# v", pretty.Formatter(actFields))
	})
}

func TestField_Path(t *testing.T) {
	f := looker.Field{
		Name:       "Bar",
		ImportPath: looker.ImportElement{},
		BaseType:   "string",
		UserType:   "string",
		Anonymous:  false,
		Tag:        "",
		Parents:    []string{"Lvl22", "Lvl3"},
	}
	assert.Equal(t, "Lvl22.Lvl3.Bar", f.Path())
	f.Parents = []string{}
	assert.Equal(t, "Bar", f.Path())
}

func TestIsProcessRower(t *testing.T) {
	for _, tc := range []struct {
		typ reflect.Type
		exp bool
	}{
		{reflect.TypeOf(testdata.Req1{}), false},
		{reflect.TypeOf(&testdata.Req1{}), false},
		{reflect.TypeOf(testdata.Req2{}), true},
		{reflect.TypeOf(&testdata.Req2{}), true},
	} {
		var typ reflect.Type = tc.typ
		if tc.typ.Kind() == reflect.Ptr {
			typ = tc.typ.Elem()
		}
		assert.Equal(t, tc.exp, looker.IsProcessRower(reflect.New(typ).Interface()), "input typ %q", typ.String())
	}
}
