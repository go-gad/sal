package looker

import (
	"log"
	"reflect"
)

func LookAt(typ reflect.Type) {
	pf("start analyze pkg %q interface %q", typ.PkgPath(), typ.Name())
}

func p(kv ...interface{}) {
	log.Print(kv...)
}

func pf(s string, kv ...interface{}) {
	log.Printf(s, kv...)
}
