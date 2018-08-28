package looker_test

import (
	"io/ioutil"
	"os"
	"testing"

	"bytes"
	"encoding/gob"

	"github.com/go-gad/sal/looker"
	"github.com/kr/pretty"
)

func TestReflect(t *testing.T) {
	pkg, err := looker.Reflect("github.com/go-gad/sal/examples/bookstore1", []string{"Store"})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Package %# v", pretty.Formatter(pkg))
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
	pkg, err := looker.Reflect("github.com/go-gad/sal/examples/bookstore1", []string{"Store"})
	if err != nil {
		t.Fatal(err)
	}

	if err := looker.EncodeGob(filename, pkg); err != nil {
		t.Fatal(err)
	}

	fb, _ := ioutil.ReadFile(filename)
	t.Logf("File content:\n%s", string(fb))

	gb := bytes.NewBuffer(fb)
	var pkgD looker.Package
	if err := gob.NewDecoder(gb).Decode(&pkgD); err != nil {
		t.Fatal(err)
	}
	t.Logf("Package %# v", pretty.Formatter(pkg))
}
