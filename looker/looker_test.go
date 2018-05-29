package looker_test

import (
	"reflect"
	"testing"

	"io/ioutil"
	"os"

	"github.com/go-gad/sal/looker"
	pkg_ "github.com/go-gad/sal/looker/bookstore"
)

func TestLookAtInterfaces(t *testing.T) {
	pf := getLogger(t)
	pkgPath := "github.com/go-gad/sal/looker/bookstore"
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
			if prm.BaseType == "struct" {
				for _, f := range prm.Fields {
					pf("\t\t\tfield %#v", f)
				}
			}
		}
		for _, prm := range v.Out {
			pf("\t\tparam OUT %#v", prm)
			if prm.BaseType == "struct" {
				for _, f := range prm.Fields {
					pf("\t\t\tfield %#v", f)
				}
			}
		}
	}

}

func TestEncodeGob(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	t.Log("filename ", filename)
	defer os.Remove(filename)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	pkg := &looker.Package{PkgPath: "some/path"}

	if err := looker.EncodeGob(filename, pkg); err != nil {
		t.Fatal(err)
	}

	fb, _ := ioutil.ReadFile(filename)
	t.Logf("File content:\n%s", string(fb))
}

func getLogger(t *testing.T) func(string, ...interface{}) {
	return func(s string, kv ...interface{}) {
		t.Logf(s, kv...)
	}
}
