package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"

	"strings"

	"github.com/go-gad/sal/looker"
)

const (
	Prefix = "Sal"
)

type generator struct {
	buf    bytes.Buffer
	indent string
}

func (g *generator) p(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, g.indent+format+"\n", args...)
}

func (g *generator) in() {
	g.indent += "\t"
}

func (g *generator) out() {
	if len(g.indent) > 0 {
		g.indent = g.indent[0 : len(g.indent)-1]
	}
}

func (g *generator) Generate(pkg *looker.Package, pkgName string) error {
	g.p("// Code generated by SalGen. DO NOT EDIT.")
	g.p("package %v", pkgName)

	g.p("import (")
	g.p("%q", "context")
	g.p("%q", pkg.ImportPath.Path)
	g.p("%q", "github.com/pkg/errors")
	g.p(")")

	for _, intf := range pkg.Interfaces {
		if err := g.GenerateInterface(intf); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) GenerateInterface(intf *looker.Interface) error {
	implName := Prefix + intf.Name
	g.p("type %v struct {", implName)
	g.p("DB *sql.DB")
	g.p("}")

	g.p("func New%v(db *sql.DB) *%v {", intf.Name, implName)
	g.p("return &%v{DB: db}", implName)
	g.p("}")

	for _, mtd := range intf.Methods {
		if err := g.GenerateMethod(implName, mtd); err != nil {
			return err
		}
	}

	return nil
}

type prmArgs []string

func (pa prmArgs) String() string {
	return strings.Join(pa, ",")
}

func (g *generator) GenerateMethod(implName string, mtd *looker.Method) error {
	g.p("")

	inArgs := make(prmArgs, 0, 2)
	inArgs = append(inArgs, "ctx "+mtd.In[0].String())
	req := mtd.In[1]
	inArgs = append(inArgs, "req "+req.String())

	// todo: array type
	outArgs := make(prmArgs, 0, 2)

	resp := mtd.Out[0]
	outArgs = append(outArgs, resp.String())

	outArgs = append(outArgs, "error")

	g.p("func (s *%v) %v(%v) (%v) {", implName, mtd.Name, inArgs.String(), outArgs.String())
	g.p("var reqMap = make(sal.KeysIntf)")
	//fmt.Printf("%# v", pretty.Formatter(req))
	for _, field := range req.Fields() {
		g.p("reqMap[%q] = &req.%s", field.Name, field.Name)
	}
	g.p("pgQuery, args := processQueryAndArgs(req.Query(), reqMap)")

	g.p("rows, err := s.DB.Query(pgQuery, args...)")
	g.ifErr("failed to execute Query")
	g.p("defer rows.Close()")
	g.p("cols, err := rows.Columns()")
	g.ifErr("failed to fetch columns")

	g.p("if !rows.Next() {")
	g.p("if err := rows.Err(); err != nil {")
	g.p("return nil, errors.Wrap(err, %q)", "rows error")
	g.p("}")
	g.p("return nil, sql.ErrNoRows")
	g.p("}")

	g.p("var resp %s", resp.String())
	g.p("var mm = make(sal.KeysIntf)")
	for _, field := range resp.Fields() {
		g.p("mm[%q] = &resp.%s", field.Name, field.Name)
	}
	g.p("var dest = make([]interface{}, 0, len(mm))")
	g.p("for _, v := range cols {")
	g.p("if intr, ok := mm[v]; ok {")
	g.p("dest = append(dest, intr)")
	g.p("}")
	g.p("}")

	g.p("if err = rows.Scan(dest...); err != nil {")
	g.p("return nil, errors.Wrap(err, %q)", "failed to scan row")
	g.p("}")

	g.p("return &resp, nil")
	g.p("}")

	return nil
}

func (g *generator) ifErr(msg string) {
	g.p("if err != nil {")
	g.p("return nil, errors.Wrap(err, %q)", msg)
	g.p("}")
}

// Output returns the generator's output, formatted in the standard Go style.
func (g *generator) Output() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format generated source code: %s\n%s", err, g.buf.String())
	}
	return src
}