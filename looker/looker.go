package looker

import (
	"path"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-gad/sal"
)

type Package struct {
	ImportPath ImportElement
	Interfaces Interfaces
}

func (p *Package) ImportPaths() []string {
	list := make([]string, 0)
	if p.ImportPath.Path != "" {
		list = append(list, p.ImportPath.Path)
	}
	for _, intf := range p.Interfaces {
		list = append(list, intf.ImportPaths()...)
	}
	return list
}

// ImportElement represents the imported package.
// Attribute `Alias` represents the optional alias for the package.
//		import foo "github.com/fooooo/baaaar-pppkkkkggg"
type ImportElement struct {
	Path  string
	Alias string
}

func (ie ImportElement) Name() string {
	if ie.Alias != "" {
		return ie.Alias
	}
	if ie.Path == "" {
		return ""
	}

	return path.Base(ie.Path)
}

func LookAtInterfaces(pkgPath string, is []reflect.Type) *Package {
	pkg := Package{
		ImportPath: ImportElement{Path: pkgPath},
		Interfaces: make(Interfaces, 0, len(is)),
	}
	for _, it := range is {
		intf := LookAtInterface(it)
		pkg.Interfaces = append(pkg.Interfaces, intf)
	}

	return &pkg
}

type Interface struct {
	ImportPath ImportElement
	UserType   string
	Methods    Methods
}

func (intf *Interface) Name(dstPath string) string {
	if dstPath == intf.ImportPath.Path {
		return intf.UserType
	}
	return intf.ImportPath.Name() + "." + intf.UserType
}

func (intf *Interface) ImplementationName(prefix string) string {
	return prefix + intf.UserType
}

func (intf *Interface) ImportPaths() []string {
	list := make([]string, 0)
	if intf.ImportPath.Path != "" {
		list = append(list, intf.ImportPath.Path)
	}

	for _, m := range intf.Methods {
		if m.Name == "Tx" || m.Name == "BeginTx" {
			continue
		}
		list = append(list, m.ImportPaths()...)
	}
	return list
}

type Interfaces []*Interface

func LookAtInterface(typ reflect.Type) *Interface {
	intf := &Interface{
		ImportPath: ImportElement{Path: typ.PkgPath()},
		UserType:   typ.Name(),
		Methods:    make(Methods, 0, typ.NumMethod()),
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

type Method struct {
	Name string
	In   Parameters
	Out  Parameters
}

func (m *Method) ImportPaths() []string {
	list := make([]string, 0)
	for _, prm := range m.In {
		list = append(list, prm.ImportPaths()...)
	}
	for _, prm := range m.Out {
		list = append(list, prm.ImportPaths()...)
	}
	return list
}

type Methods []*Method

type Parameter interface {
	Kind() string
	Name(dstPath string) string
	Pointer() bool
	ImportPaths() []string
}

type Parameters []Parameter

func LookAtFuncParameters(mt reflect.Type) (Parameters, Parameters) {
	var in = make(Parameters, 0)
	for i := 0; i < mt.NumIn(); i++ {
		in = append(in, LookAtParameter(mt.In(i)))
	}

	var out = make(Parameters, 0)
	for i := 0; i < mt.NumOut(); i++ {
		out = append(out, LookAtParameter(mt.Out(i)))
	}

	return in, out
}

// Use exported fields because god.Encoder
type StructElement struct {
	ImportPath   ImportElement
	UserType     string
	IsPointer    bool
	Fields       Fields
	ProcessRower bool
}

func (prm *StructElement) Kind() string {
	return reflect.Struct.String()
}

func (prm *StructElement) Name(dstPath string) string {
	if dstPath == prm.ImportPath.Path {
		return prm.UserType
	}
	return prm.ImportPath.Name() + "." + prm.UserType
}

func (prm *StructElement) Pointer() bool {
	return prm.IsPointer
}

// todo: import path for fields
func (prm *StructElement) ImportPaths() []string {
	if prm.ImportPath.Path != "" {
		return []string{prm.ImportPath.Path}
	}
	return []string{}
}

type SliceElement struct {
	ImportPath ImportElement
	UserType   string
	Item       Parameter
	IsPointer  bool
}

func (prm *SliceElement) Kind() string {
	return reflect.Slice.String()
}

func (prm *SliceElement) Name(dstPath string) string {
	if prm.UserType != "" {
		if dstPath == prm.ImportPath.Path {
			return prm.UserType
		}
		return prm.ImportPath.Name() + "." + prm.UserType
	}

	var ptr string
	if prm.Item.Pointer() {
		ptr = "*"
	}
	return "[]" + ptr + prm.Item.Name(dstPath)
}

func (prm *SliceElement) Pointer() bool {
	return prm.IsPointer
}

func (prm *SliceElement) ImportPaths() []string {
	if prm.ImportPath.Path != "" {
		return []string{prm.ImportPath.Path}
	}
	return []string{}
}

type InterfaceElement struct {
	ImportPath ImportElement
	UserType   string
}

func (prm *InterfaceElement) Kind() string {
	return reflect.Interface.String()
}

func (prm *InterfaceElement) Name(dstPath string) string {
	if prm.ImportPath.Path == "" {
		return prm.UserType
	}
	if dstPath == prm.ImportPath.Path {
		return prm.UserType
	}
	return prm.ImportPath.Name() + "." + prm.UserType
}

func (prm *InterfaceElement) Pointer() bool {
	return false
}

func (prm *InterfaceElement) ImportPaths() []string {
	if prm.ImportPath.Path != "" {
		return []string{prm.ImportPath.Path}
	}
	return []string{}
}

type UnsupportedElement struct {
	ImportPath ImportElement
	UserType   string
	BaseType   string
	IsPointer  bool
}

func (prm *UnsupportedElement) Kind() string {
	return prm.BaseType
}

func (prm *UnsupportedElement) Name(dstPath string) string {
	if dstPath == prm.ImportPath.Path {
		return prm.UserType
	}
	return prm.ImportPath.Name() + "." + prm.UserType
}

func (prm *UnsupportedElement) Pointer() bool {
	return prm.IsPointer
}

func (prm *UnsupportedElement) ImportPaths() []string {
	if prm.ImportPath.Path != "" {
		return []string{prm.ImportPath.Path}
	}
	return []string{}
}

func LookAtParameter(at reflect.Type) Parameter {
	var pointer bool
	if at.Kind() == reflect.Ptr {
		at = at.Elem()
		pointer = true
	}
	var prm Parameter

	im := GetImportElement(at)

	switch at.Kind() {
	case reflect.Struct:
		prm = &StructElement{
			ImportPath:   im,
			UserType:     at.Name(),
			IsPointer:    pointer,
			Fields:       LookAtFields(at),
			ProcessRower: IsProcessRower(reflect.New(at).Interface()),
		}
	case reflect.Slice:
		prm = &SliceElement{
			ImportPath: im,
			UserType:   at.Name(),
			IsPointer:  pointer,
			Item:       LookAtParameter(at.Elem()),
		}
	case reflect.Interface:
		prm = &InterfaceElement{
			ImportPath: im,
			UserType:   at.Name(),
		}
	default:
		prm = &UnsupportedElement{
			ImportPath: im,
			UserType:   at.Name(),
			BaseType:   at.Kind().String(),
			IsPointer:  pointer,
		}
	}

	return prm
}

func IsProcessRower(s interface{}) bool {
	_, ok := s.(sal.ProcessRower)

	return ok
}

// Field describes the fields of struct after reflection.
type Field struct {
	// See the fields that describe Req struct.
	// type Req struct {
	//	ID   int64 `sql:"id"`
	// }
	// for Req.ID Name contains `ID`.
	Name string
	// ImportPath contains ImportElement.
	ImportPath ImportElement
	// for Req.ID BaseType contains `int64`.
	BaseType string
	// UserType contains type other then basic if it's defined otherwise basic.
	UserType string
	// Anonymous sets to true if field contains anonymous nested struct.
	Anonymous bool
	// Tag contains the value for tag with name `sql` if it's presented.
	Tag string
	// todo
	Parents []string
}

// ColumnName returns the column name to use for mapping with sql response.
func (f Field) ColumnName() string {
	if f.Tag == "" {
		return f.Name
	}
	return f.Tag
}

func (f Field) Path() string {
	path := append(f.Parents, f.Name)
	return strings.Join(path, ".")
}

// Fields is alias for slice of Field.
type Fields []Field

// tagName contains the name of tag of struct field to make mapping with sql response.
const tagName = "sql"

// LookAtFields receives the reflect.Type object of struct and returns the Fields.
func LookAtFields(st reflect.Type) Fields {
	fields := make(Fields, 0, st.NumField())
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		fields = append(fields, LookAtField(ft)...)
	}
	return fields
}

// LookAtField receive the reflected object of struct field and return Fields.
// If field points to anonymous struct then LookAtFields will be called.
func LookAtField(ft reflect.StructField) Fields {
	if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
		// going to analyze embedded struct
		list := LookAtFields(ft.Type)
		for i := range list {
			list[i].Parents = append([]string{ft.Name}, list[i].Parents...)
		}
		return list
	}
	f := Field{
		Name:       ft.Name,
		ImportPath: ImportElement{Path: ft.Type.PkgPath()},
		BaseType:   ft.Type.Kind().String(),
		UserType:   ft.Type.Name(),
		Anonymous:  ft.Anonymous,
		Tag:        ft.Tag.Get(tagName),
		Parents:    make([]string, 0),
	}
	return []Field{f}
}

func GetImportElement(typ reflect.Type) ImportElement {
	alias := getAlias(typ.String())
	im := ImportElement{Path: typ.PkgPath()}
	if alias != "" && im.Name() != alias {
		im.Alias = alias
	}
	return im
}

// return on []*foo.Body the string foo
func getAlias(str string) string {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	ar := strings.FieldsFunc(str, f)
	if len(ar) == 2 {
		return ar[0]
	}
	return ""
}
