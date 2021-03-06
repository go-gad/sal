package main

import (
	"io/ioutil"
	"testing"

	"github.com/go-gad/sal/looker"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var update bool = false

func TestGenerateCode(t *testing.T) {
	dstPkg := looker.ImportElement{Path: "github.com/go-gad/sal/examples/bookstore"}
	code, err := GenerateCode(dstPkg, "github.com/go-gad/sal/examples/bookstore", []string{"Store"})
	if err != nil {
		t.Fatalf("Failed to generate a code: %+v", err)
	}

	//t.Logf("\n%s", string(code))
	if update {
		if err = ioutil.WriteFile("../examples/bookstore/sal_client.go", code, 0666); err != nil {
			t.Fatalf("failed to write file: %+v", err)
		}
	}
	expCode, err := ioutil.ReadFile("../examples/bookstore/sal_client.go")
	if string(expCode) != string(code) {
		t.Error("generated code is not equal to expected")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expCode), string(code), true)
		t.Log(dmp.DiffPrettyText(diffs))
	}
}

func TestGenerateCode2(t *testing.T) {
	dstPkg := looker.ImportElement{Path: "github.com/go-gad/sal/looker/testdata"}
	code, err := GenerateCode(dstPkg, "github.com/go-gad/sal/looker/testdata", []string{"Store"})
	if err != nil {
		t.Fatalf("Failed to generate a code: %+v", err)
	}

	t.Logf("\n%s", string(code))

	if update {
		if err = ioutil.WriteFile("../looker/testdata/store_client.go", code, 0666); err != nil {
			t.Fatalf("failed to write file: %+v", err)
		}
	}
}

func TestGenerator_GenerateRowMap(t *testing.T) {
	prm := &looker.UnsupportedElement{
		ImportPath: looker.ImportElement{},
		UserType:   "int64",
		BaseType:   "int64",
		IsPointer:  true,
	}
	g := new(generator)
	if err := g.GenerateRowMap(prm, "rowMap", "resp"); err == nil {
		t.Error("should be error")
	}
}
