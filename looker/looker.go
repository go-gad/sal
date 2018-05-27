package looker

import (
	"log"
	"reflect"
)

func LookAtInterface(typ reflect.Type) *Interface {
	//pf("start analyze pkg %q interface %q", typ.PkgPath(), typ.Name())
	//pkg := &Package{
	//Name: path.Base(typ.PkgPath()),
	//}
	//pf("%#v", pkg)

	intf := &Interface{
		Name:    typ.Name(),
		Methods: make(Methods, 0, typ.NumMethod()),
	}
	//pkg.Interface = intf
	//pf("%#v", intf)

	//p("-------")
	for i := 0; i < typ.NumMethod(); i++ {
		mt := typ.Method(i)
		m := Method{
			Name: mt.Name,
		}
		//pf("%#v", m)
		in := LookAtFuncParameters(typ.Method(i).Type)
		m.In = in
		//p("-------")
		intf.Methods = append(intf.Methods, &m)
	}
	return intf
}

func LookAtFuncParameters(mt reflect.Type) Parameters {
	//pf("look at args for kind %q", mt.Kind())
	var in = make([]*Parameter, 0)
	for i := 0; i < mt.NumIn(); i++ {
		in = append(in, LookAtParameter(mt.In(i)))
	}

	return in
}

func LookAtParameter(at reflect.Type) *Parameter {
	var pointer bool
	if at.Kind() == reflect.Ptr {
		at = at.Elem()
		pointer = true
	}
	return &Parameter{
		Pkg:      at.PkgPath(),
		BaseType: at.Kind().String(),
		UserType: at.Name(),
		Pointer:  pointer,
	}
	//
	//switch at.Kind() {
	//case reflect.Interface:
	//	pf("parameter name %q type %q basepkg %q", at.Name(), at.Kind(), path.Base(at.PkgPath()))
	//case reflect.Ptr:
	//	at = at.Elem()
	//	pf("parameter name %q type %q basepkg %q", at.Name(), at.Kind(), path.Base(at.PkgPath()))
	//	LookAtFields(at)
	//default:
	//	pf("unsupported parameter name %q type %q basepkg %q", at.Name(), at.Kind(), path.Base(at.PkgPath()))
	//}
	//
	//return &prm
}

func LookAtFields(st reflect.Type) {
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		pf(">>> field %s", ft.Name)
	}
}

type Package struct {
	Name      string
	Interface *Interface
}

type Interface struct {
	Name    string
	Methods Methods
}

type Method struct {
	Name string
	In   Parameters
}

type Methods []*Method

type Parameter struct {
	Pkg      string
	BaseType string
	UserType string
	Pointer  bool
}

type Parameters []*Parameter

func p(kv ...interface{}) {
	log.Print(kv...)
}

func pf(s string, kv ...interface{}) {
	log.Printf(s, kv...)
}
