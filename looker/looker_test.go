package looker_test

import (
	"reflect"
	"testing"

	"github.com/go-gad/sal/looker"
	"github.com/go-gad/sal/looker/bookstore"
)

func TestLookAtInterface(t *testing.T) {
	var typ reflect.Type = reflect.TypeOf((*bookstore.StoreClient)(nil)).Elem()
	looker.LookAtInterface(typ)
}
