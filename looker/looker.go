package looker

import (
	"log"
	"path"
	"reflect"
)

func LookAtInterfaces(pkgPath string, is []reflect.Type) *Package {
	pkg := Package{
		ImportPath: ImportElement{Path:pkgPath},
		Interfaces: make(Interfaces, 0, len(is)),
	}
	for _, it := range is {
		intf := LookAtInterface(it)
		pkg.Interfaces = append(pkg.Interfaces, intf)
	}

	return &pkg
}

func LookAtInterface(typ reflect.Type) *Interface {
	intf := &Interface{
		Name:    typ.Name(),
		Methods: make(Methods, 0, typ.NumMethod()),
	}

	for i := 0; i < typ.NumMethod(); i++ {
		mt := typ.Method(i)
		m := Method{
			Name: mt.Name,
		}
		in, out := LookAtFuncParameters(typ.Method(i).Type)
		m.In = in
		m.Out = out

		intf.Methods = append(intf.Methods, &m)
	}
	return intf
}

func LookAtFuncParameters(mt reflect.Type) ([]*ParameterStruct, []*ParameterStruct) {
	var in = make([]*ParameterStruct, 0)
	for i := 0; i < mt.NumIn(); i++ {
		in = append(in, LookAtParameter(mt.In(i)))
	}

	var out = make([]*ParameterStruct, 0)
	for i := 0; i < mt.NumOut(); i++ {
		out = append(out, LookAtParameter(mt.Out(i)))
	}

	return in, out
}

func LookAtParameter(at reflect.Type) *ParameterStruct {
	var pointer bool
	if at.Kind() == reflect.Ptr {
		at = at.Elem()
		pointer = true
	}
	prm := ParameterStruct{
		ImportPath: at.PkgPath(),
		BaseType:   at.Kind().String(),
		UserType:   at.Name(),
		Pointer:    pointer,
	}

	if prm.BaseType == reflect.Struct.String() {
		prm.Fields = LookAtFields(at)
	}

	return &prm
}

func LookAtFields(st reflect.Type) Fields {
	fields := make(Fields, 0, st.NumField())
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		field := &Field{
			Name:       ft.Name,
			ImportPath: ft.Type.PkgPath(),
			BaseType:   ft.Type.Kind().String(),
			UserType:   ft.Type.Name(),
			Anonymous:  ft.Anonymous,
		}
		fields = append(fields, field)
	}
	return fields
}

type Package struct {
	ImportPath ImportElement
	Interfaces Interfaces
}

// ImportElement represents the imported package.
// Attribute `Alias` represents the optional alias for the package.
//		import foo "github.com/fooooo/baaaar-pppkkkkggg"
type ImportElement struct {
	Path string
	Alias string
}

func (ie ImportElement) Name() string {
	if ie.Alias != "" {
		return ie.Alias
	}

	return path.Base(ie.Path)
}

type Interface struct {
	Name    string
	Methods Methods
}

type Interfaces []*Interface

type Method struct {
	Name string
	In   []*ParameterStruct
	Out  []*ParameterStruct
}

type Methods []*Method

const (
	ParameterTypeStruct = "struct"
)

type Parameter interface {
	Type() string
	String() string
}

type Parameters []Parameter

type ParameterStruct struct {
	ImportPath string
	BaseType   string
	UserType   string
	Pointer    bool
	Fields     Fields
}

func (prm *ParameterStruct) Type() string {
	return ParameterTypeStruct
}

func (prm *ParameterStruct) PkgAlias() string {
	return path.Base(prm.ImportPath)
}

func (prm *ParameterStruct) PtrPrefix() string {
	if prm.Pointer {
		return "*"
	}
	return ""
}

func (prm *ParameterStruct) String() string {
	return prm.PtrPrefix() + prm.PkgAlias() + "." + prm.UserType
}

type Field struct {
	Name       string
	ImportPath string
	BaseType   string
	UserType   string
	Anonymous  bool
}

type Fields []*Field

func p(kv ...interface{}) {
	log.Print(kv...)
}

func pf(s string, kv ...interface{}) {
	log.Printf(s, kv...)
}
