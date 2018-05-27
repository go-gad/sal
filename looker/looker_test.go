package looker_test

import (
	"reflect"
	"testing"

	"github.com/go-gad/sal/looker"
	"github.com/go-gad/sal/looker/bookstore"
)

func TestLookAtInterface(t *testing.T) {
	pf := getLogger(t)
	var typ reflect.Type = reflect.TypeOf((*bookstore.StoreClient)(nil)).Elem()
	intf := looker.LookAtInterface(typ)
	pf("Interface %#v", intf)
	for i, v := range intf.Methods {
		pf("\t[%d] method %q", i, v.Name)
		for _, prm := range v.In {
			pf("\t\tparam %#v", prm)
		}
	}

}

func getLogger(t *testing.T) func(string, ...interface{}) {
	return func(s string, kv ...interface{}) {
		t.Logf(s, kv...)
	}
}
