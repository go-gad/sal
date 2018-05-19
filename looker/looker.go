package looker

import (
	"log"
	"path"
	"reflect"
)

func LookAtInterface(typ reflect.Type) {
	pf("start analyze pkg %q interface %q", typ.PkgPath(), typ.Name())
	pkg := &Package{
		Name: path.Base(typ.PkgPath()),
	}
	pf("%#v", pkg)

	intf := &Interface{
		Name:    typ.Name(),
		Methods: make(Methods, 0, typ.NumMethod()),
	}
	pkg.Interface = intf
	pf("%#v", intf)

	p("-------")
	for i := 0; i < typ.NumMethod(); i++ {
		mt := typ.Method(i)
		m := &Method{
			Name: mt.Name,
		}
		pf("%#v", m)
		LookAtMethodParameters(typ.Method(i).Type)
		p("-------")
	}
}

func LookAtMethodParameters(mt reflect.Type) {
	pf("look at args for %q", mt.Kind())
	for i := 0; i < mt.NumIn(); i++ {
		LookAtParameter(mt.In(i))
	}

}

func LookAtParameter(at reflect.Type) {
	switch at.Kind() {
	case reflect.Interface:
		pf("parameter name %q type %q basepkg %q", at.Name(), at.Kind(), path.Base(at.PkgPath()))
	case reflect.Ptr:
		at = at.Elem()
		pf("parameter name %q type %q basepkg %q", at.Name(), at.Kind(), path.Base(at.PkgPath()))
	default:
		pf("unsupported parameter name %q type %q basepkg %q", at.Name(), at.Kind(), path.Base(at.PkgPath()))
	}
}

type Package struct {
	Name      string
	Interface *Interface
}

type Interface struct {
	Name    string
	Methods []*Method
}

type Method struct {
	Name string
}

type Methods []*Method

type Parameter struct {
	Name string
}

func p(kv ...interface{}) {
	log.Print(kv...)
}

func pf(s string, kv ...interface{}) {
	log.Printf(s, kv...)
}
