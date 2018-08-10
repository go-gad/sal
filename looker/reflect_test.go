package looker_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-gad/sal/looker"
	"github.com/kr/pretty"
	"bytes"
	"encoding/gob"
)

func TestReflect(t *testing.T) {
	pkg, err := looker.Reflect("github.com/go-gad/sal/examples/bookstore1", []string{"StoreClient"})
	if err != nil {
		t.Fatal(err)
	}
	pf := getLogger(t)
	pf("Package %# v", pretty.Formatter(pkg))
}

func TestEncodeGob(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	//t.Log("filename ", filename)
	defer os.Remove(filename)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	pkg := &looker.Package{ImportPath: looker.ImportElement{Path:"some/path"}}

	if err := looker.EncodeGob(filename, pkg); err != nil {
		t.Fatal(err)
	}

	fb, _ := ioutil.ReadFile(filename)
	//t.Logf("File content:\n%s", string(fb))

	gb := bytes.NewBuffer(fb)
	var pkgD looker.Package
	if err := gob.NewDecoder(gb).Decode(&pkgD); err != nil {
		t.Fatal(err)
	}
}
