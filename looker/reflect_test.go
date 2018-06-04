package looker_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-gad/sal/looker"
)

func TestReflect(t *testing.T) {
	pkg, err := looker.Reflect("github.com/go-gad/sal/internal/bookstore", []string{"StoreClient"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Package %#v", pkg)
	pf := getLogger(t)
	for _, intf := range pkg.Interfaces {
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
	pkg := &looker.Package{ImportPath: "some/path"}

	if err := looker.EncodeGob(filename, pkg); err != nil {
		t.Fatal(err)
	}

	fb, _ := ioutil.ReadFile(filename)
	t.Logf("File content:\n%s", string(fb))
}
