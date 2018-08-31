package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"strings"

	"reflect"

	"github.com/go-gad/sal"
	"github.com/go-gad/sal/looker"
	"github.com/pkg/errors"
)

const (
	Prefix = "Sal"
)

const (
	MethodNameTx      string = "Tx"
	MethodNameBeginTx string = "BeginTx"
)

type generator struct {
	buf    bytes.Buffer
	indent string
}

func (g *generator) Generate(pkg *looker.Package, dstPkg looker.ImportElement) error {
	g.p("// Code generated by SalGen. DO NOT EDIT.")
	//g.p("// Generated at %s", time.Now())
	g.p("package %v", dstPkg.Name())

	g.p("import (")
	g.p("%q", "context")
	g.p("%q", "database/sql")
	g.br()
	if pkg.ImportPath.Path != dstPkg.Path {
		g.p("%q", pkg.ImportPath.Path)
	}
	g.p("%q", "github.com/pkg/errors")
	g.p("%q", "github.com/go-gad/sal")

	g.p(")")

	for _, intf := range pkg.Interfaces {
		if err := g.GenerateInterface(dstPkg, intf); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) GenerateInterface(dstPkg looker.ImportElement, intf *looker.Interface) error {
	g.p("type %vConn interface {", intf.UserType)
	g.p("%v", intf.UserType)
	g.br()
	g.p("BeginTx(opts *sql.TxOptions) (%vTx, error)", intf.UserType)
	g.p("}")
	g.br()

	g.p("type %vTx interface {", intf.UserType)
	g.p("%v", intf.UserType)
	g.br()
	g.p("Commit() error")
	g.p("Rollback() error")
	g.p("}")
	g.br()

	implName := intf.ImplementationName(dstPkg.Path, Prefix)
	g.p("type %v struct {", implName)
	g.p("handler sal.QueryHandler")
	g.p("ctrl *sal.Controller")
	g.p("txOpened bool")
	g.p("}")

	g.p("func New%v(h sal.QueryHandler, options ...sal.ClientOption) *%v {", intf.UserType, implName)
	g.p("s := &%s{", implName)
	g.p("handler: h,")
	g.p("ctrl: sal.NewController(options...),")
	g.p("txOpened: false,")
	g.p("}")
	g.br()
	g.p("return s")
	g.p("}")
	g.br()

	g.GenerateBeginTx(dstPkg, intf)
	g.br()
	g.GenerateTx(dstPkg, intf)

	for _, mtd := range intf.Methods {
		if err := g.GenerateMethod(dstPkg, implName, mtd); err != nil {
			return err
		}
		g.br()
	}

	return nil
}

type prmArgs []string

func (pa prmArgs) String() string {
	return strings.Join(pa, ",")
}

func (g *generator) GenerateMethod(dstPkg looker.ImportElement, implName string, mtd *looker.Method) error {
	switch mtd.Name {
	case MethodNameBeginTx, MethodNameTx:
		return nil
	}

	inArgs := make(prmArgs, 0, 2)
	inArgs = append(inArgs, "ctx "+mtd.In[0].Name(dstPkg.Path))
	req := mtd.In[1]
	inArgs = append(inArgs, "req "+elementType(req.Pointer(), req.Name(dstPkg.Path)))

	operation := calcOperationType(mtd.Out)

	outArgs := make(prmArgs, 0, 2)

	resp := mtd.Out[0]
	if operation != sal.OperationTypeExec {
		outArgs = append(outArgs, elementType(resp.Pointer(), resp.Name(dstPkg.Path)))
	}
	outArgs = append(outArgs, mtd.Out[len(mtd.Out)-1].Name(dstPkg.Path))

	g.p("func (s *%v) %v(%v) (%v) {", implName, mtd.Name, inArgs.String(), outArgs.String())
	g.p("var (")
	g.p("err error")
	g.p("rawQuery = req.Query()")
	g.p("reqMap = make(sal.RowMap)")
	g.p(")")
	if req.Kind() == reflect.Struct.String() {
		reqSt := req.(*looker.StructElement)
		for _, field := range reqSt.Fields {
			g.p("reqMap[%q] = &req.%s", field.ColumnName(), field.Name)
		}
		g.br()
		if reqSt.ProcessRower {
			g.p("req.ProcessRow(reqMap)")
			g.br()
		}
	} else {
		return errors.New("unsupported type of request variable")
	}

	g.p("ctx = context.WithValue(ctx, sal.ContextKeyTxOpened, s.txOpened)")
	g.p("ctx = context.WithValue(ctx, sal.ContextKeyOperationType, %q)", operation.String())
	g.p("ctx = context.WithValue(ctx, sal.ContextKeyMethodName, %q)", mtd.Name)
	g.br()

	g.p("pgQuery, args := sal.ProcessQueryAndArgs(rawQuery, reqMap)")
	g.br()

	g.p("stmt, err := s.ctrl.PrepareStmt(ctx, s.handler, pgQuery)")
	g.p("if err != nil {")
	switch operation {
	case sal.OperationTypeQuery, sal.OperationTypeQueryRow:
		g.p("return nil, errors.WithStack(err)")
	case sal.OperationTypeExec:
		g.p("return errors.WithStack(err)")
	}
	g.p("}")
	g.br()

	g.beforeQueryHook("rawQuery", "req")
	g.br()

	switch operation {
	case sal.OperationTypeQuery, sal.OperationTypeQueryRow:
		g.p("rows, err := stmt.QueryContext(ctx, args...)")
		g.ifErr("failed to execute Query")
		g.p("defer rows.Close()")
		g.br()

		g.p("cols, err := rows.Columns()")
		g.ifErr("failed to fetch columns")
		g.br()
	case sal.OperationTypeExec:
		g.p("_, err = stmt.ExecContext(ctx, args...)")
		g.p("if err != nil {")
		g.p("return errors.Wrap(err, %q)", "failed to execute Exec")
		g.p("}")
		g.br()
	}

	if operation == sal.OperationTypeExec {
		g.p("return nil")
		g.p("}")
		return nil
	}

	if operation == sal.OperationTypeQueryRow {
		g.p("if !rows.Next() {")
		g.p("if err = rows.Err(); err != nil {")
		g.p("return nil, errors.Wrap(err, %q)", "rows error")
		g.p("}")
		g.p("return nil, sql.ErrNoRows")
		g.p("}")
		g.br()
	}

	var respRow looker.Parameter
	if operation == sal.OperationTypeQuery {
		g.p("var list = make(%s, 0)", resp.Name(dstPkg.Path))
		g.br()
		g.p("for rows.Next() {")
		respSlice := resp.(*looker.SliceElement)

		respRow = respSlice.Item
	} else {
		respRow = resp
	}
	var respRowStr = "resp"
	g.p("var %s %s", respRowStr, respRow.Name(dstPkg.Path))
	g.p("var respMap = make(sal.RowMap)")
	if respRow.Kind() == reflect.Struct.String() {
		respSt := respRow.(*looker.StructElement)
		for _, field := range respSt.Fields {
			g.p("respMap[%q] = &resp.%s", field.ColumnName(), field.Name)
		}
		g.br()
		if respSt.ProcessRower {
			g.p("%s.ProcessRow(respMap)", respRowStr)
			g.br()
		}
	}
	g.p("var dest = make([]interface{}, 0, len(respMap))")
	g.p("for _, v := range cols {")
	g.p("if intr, ok := respMap[v]; ok {")
	g.p("dest = append(dest, intr)")
	g.p("}")
	g.p("}")
	g.br()

	g.p("if err = rows.Scan(dest...); err != nil {")
	g.p("return nil, errors.Wrap(err, %q)", "failed to scan row")
	g.p("}")
	if operation == sal.OperationTypeQuery {
		if respRow.Pointer() {
			respRowStr = "&resp"
		}
		g.br()
		g.p("list = append(list, %s)", respRowStr)
		g.p("}")
	}
	g.br()

	g.p("if err = rows.Err(); err != nil {")
	g.p("return nil, errors.Wrap(err, %q)", "something failed during iteration")
	g.p("}")
	g.br()

	respStr := "resp"
	if operation == sal.OperationTypeQuery {
		respStr = "list"
	}

	if resp.Pointer() {
		respStr = "&" + respStr
	}

	g.p("return %s, nil", respStr)
	g.p("}")

	return nil
}

func (g *generator) GenerateBeginTx(dstPkg looker.ImportElement, intf *looker.Interface) {
	g.p("func (s *%s) BeginTx(ctx context.Context, opts *sql.TxOptions) (*%s, error) {", intf.ImplementationName(dstPkg.Path, Prefix), intf.ImplementationName(dstPkg.Path, Prefix))
	g.p("dbConn, ok := s.handler.(sal.TransactionBegin)")
	g.p("if !ok {")
	g.p("return nil, errors.New(%q)", "oops")
	g.p("}")
	g.p("var (")
	g.p("err error")
	g.p("tx  *sql.Tx")
	g.p(")")
	g.br()

	g.p("ctx = context.WithValue(ctx, sal.ContextKeyTxOpened, s.txOpened)")
	g.p("ctx = context.WithValue(ctx, sal.ContextKeyOperationType, %q)", sal.OperationTypeBegin.String())
	g.p("ctx = context.WithValue(ctx, sal.ContextKeyMethodName, %q)", "BeginTx")
	g.br()

	g.beforeQueryHook(`"BEGIN"`, "nil")
	g.br()

	g.p("tx, err = dbConn.BeginTx(ctx, opts)")
	g.p("if err != nil {")
	g.p("err = errors.Wrap(err, %q)", "failed to start tx")
	g.p("return nil, err")
	g.p("}")
	g.br()
	g.p("newClient := &%s{", intf.ImplementationName(dstPkg.Path, Prefix))
	g.p("handler: tx,")
	g.p("ctrl: s.ctrl,")
	g.p("txOpened: true,")
	g.p("}")
	g.br()
	g.p("return newClient, nil")
	g.p("}")
}

func (g *generator) GenerateTx(dstPkg looker.ImportElement, intf *looker.Interface) {
	g.p("func (s *%s) Commit() error {", intf.ImplementationName(dstPkg.Path, Prefix))
	g.p("//todo")
	g.p("return nil")
	g.p("}")
	g.br()
	g.p("func (s *%s) Rollback() error {", intf.ImplementationName(dstPkg.Path, Prefix))
	g.p("//todo")
	g.p("return nil")
	g.p("}")
}

func (g *generator) ifErr(msg string) {
	g.p("if err != nil {")
	g.p("return nil, errors.Wrap(err, %q)", msg)
	g.p("}")
}

func (g *generator) beforeQueryHook(q, r string) {
	g.p("for _, fn := range s.ctrl.BeforeQuery {")
	g.p("var fnz sal.FinalizerFunc")
	g.p("ctx, fnz = fn(ctx, %s, %s)", q, r)
	g.p("if fnz != nil {")
	g.p("defer func() { fnz(ctx, err) }()")
	g.p("}")
	g.p("}")
}

func elementType(ptr bool, name string) string {
	var prefix string
	if ptr {
		prefix = "*"
	}
	return prefix + name
}

// Output returns the generator's output, formatted in the standard Go style.
func (g *generator) Output() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format generated source code: %s\n%s", err, g.buf.String())
	}
	return src
}

func (g *generator) p(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, g.indent+format+"\n", args...)
}

func (g *generator) br() {
	g.p("")
}

func calcOperationType(prms looker.Parameters) sal.OperationType {
	if len(prms) == 1 {
		return sal.OperationTypeExec
	}
	if prms[0].Kind() == reflect.Slice.String() {
		return sal.OperationTypeQuery
	}
	return sal.OperationTypeQueryRow
}
