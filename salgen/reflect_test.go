package main

import (
	"testing"
)

func TestReflect(t *testing.T) {
	pkg, err := Reflect("github.com/go-gad/sal/looker/bookstore", []string{"StoreClient"})
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

func getLogger(t *testing.T) func(string, ...interface{}) {
	return func(s string, kv ...interface{}) {
		t.Logf(s, kv...)
	}
}
