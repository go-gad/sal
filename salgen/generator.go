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

	paths := ImportPaths(pkg.ImportPaths(), dstPkg.Path)
	g.GenerateImportPaths(paths)

	for _, intf := range pkg.Interfaces {
		if err := g.GenerateInterface(dstPkg, intf); err != nil {
			return err
		}
		g.p("// compile time checks")
		g.p("var _ %s = &%s{}", intf.Name(dstPkg.Path), intf.ImplementationName(Prefix))
	}

	return nil
}

func (g *generator) GenerateImportPaths(paths []string) {
	g.p("import (")
	g.p("%q", "database/sql")
	for _, p := range paths {
		g.p("%q", p)
	}
	g.p("%q", "github.com/pkg/errors")
	g.p("%q", "github.com/go-gad/sal")
	g.p(")")
}

func (g *generator) GenerateInterface(dstPkg looker.ImportElement, intf *looker.Interface) error {
	implName := intf.ImplementationName(Prefix)
	g.p("type %v struct {", implName)
	g.p("%s", intf.Name(dstPkg.Path))
	g.p("handler sal.QueryHandler")
	g.p("parent sal.QueryHandler")
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
	if operation == sal.OperationTypeExec {
		if isSqlResult(resp) {
			outArgs = append(outArgs, elementType(resp.Pointer(), resp.Name(dstPkg.Path)))
		}
	} else {
		outArgs = append(outArgs, elementType(resp.Pointer(), resp.Name(dstPkg.Path)))
	}
	outArgs = append(outArgs, mtd.Out[len(mtd.Out)-1].Name(dstPkg.Path))

	g.p("func (s *%v) %v(%v) (%v) {", implName, mtd.Name, inArgs.String(), outArgs.String())
	g.p("var (")
	g.p("err error")
	g.p("rawQuery = req.Query()")
	g.p("reqMap = make(sal.RowMap)")
	g.p(")")
	g.GenerateRowMap(req, "reqMap", "req")

	g.p("ctx = context.WithValue(ctx, sal.ContextKeyTxOpened, s.txOpened)")
	g.p("ctx = context.WithValue(ctx, sal.ContextKeyOperationType, %q)", operation.String())
	g.p("ctx = context.WithValue(ctx, sal.ContextKeyMethodName, %q)", mtd.Name)
	g.br()

	g.p("pgQuery, args := sal.ProcessQueryAndArgs(rawQuery, reqMap)")
	g.br()

	var errRespStr = responseErrStr(operation, resp, dstPkg.Path)

	g.p("stmt, err := s.ctrl.PrepareStmt(ctx, s.parent, s.handler, pgQuery)")
	g.p("if err != nil {")
	switch operation {
	case sal.OperationTypeQuery, sal.OperationTypeQueryRow:
		g.p("return %s, errors.WithStack(err)", errRespStr)
	case sal.OperationTypeExec:
		if isSqlResult(mtd.Out[0]) {
			g.p("return nil, errors.WithStack(err)")
		} else {
			g.p("return errors.WithStack(err)")
		}
	}
	g.p("}")
	g.br()

	g.beforeQueryHook("rawQuery", "req")
	g.br()

	switch operation {
	case sal.OperationTypeQuery, sal.OperationTypeQueryRow:
		g.p("rows, err := stmt.QueryContext(ctx, args...)")
		g.ifErr(errRespStr, "failed to execute Query")
		g.p("defer rows.Close()")
		g.br()

		g.p("cols, err := rows.Columns()")
		g.ifErr(errRespStr, "failed to fetch columns")
		g.br()
	case sal.OperationTypeExec:
		if isSqlResult(mtd.Out[0]) {
			g.p("res, err := stmt.ExecContext(ctx, args...)")
			g.ifErr("nil", "failed to execute Exec")
		} else {
			g.p("_, err = stmt.ExecContext(ctx, args...)")
			g.ifErr("", "failed to execute Exec")
		}

		g.br()
	}

	if operation == sal.OperationTypeExec {
		if isSqlResult(mtd.Out[0]) {
			g.p("return res, nil")
		} else {
			g.p("return nil")
		}
		g.p("}")
		return nil
	}

	if operation == sal.OperationTypeQueryRow {
		g.p("if !rows.Next() {")
		g.p("if err = rows.Err(); err != nil {")
		g.p("return %s, errors.Wrap(err, %q)", errRespStr, "rows error")
		g.p("}")
		g.p("return %s, sql.ErrNoRows", errRespStr)
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
	g.GenerateRowMap(respRow, "respMap", "resp")

	g.p("dest := sal.GetDests(cols, respMap)")
	g.br()

	g.p("if err = rows.Scan(dest...); err != nil {")
	g.p("return %s, errors.Wrap(err, %q)", errRespStr, "failed to scan row")
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
	g.p("return %s, errors.Wrap(err, %q)", errRespStr, "something failed during iteration")
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

func (g *generator) GenerateRowMap(prm looker.Parameter, mapName string, prmName string) error {
	if prm.Kind() == reflect.Struct.String() {
		st := prm.(*looker.StructElement)
		for _, field := range st.Fields {
			g.p("%s.AppendTo(%q, &%s.%s)", mapName, field.ColumnName(), prmName, field.Path())
		}
		g.br()
		if st.ProcessRower {
			g.p("%s.ProcessRow(%s)", prmName, mapName)
			g.br()
		}
	} else {
		return errors.New("unsupported type of request variable")
	}

	return nil
}

func (g *generator) GenerateBeginTx(dstPkg looker.ImportElement, intf *looker.Interface) {
	g.p("func (s *%s) BeginTx(ctx context.Context, opts *sql.TxOptions) (%s, error) {", intf.ImplementationName(Prefix), intf.Name(dstPkg.Path))
	g.p("dbConn, ok := s.handler.(sal.TransactionBegin)")
	g.p("if !ok {")
	g.p("return nil, errors.New(%q)", "handler doesn't satisfy the interface TransactionBegin")
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
	g.p("newClient := &%s{", intf.ImplementationName(Prefix))
	g.p("handler: tx,")
	g.p("parent: s.handler,")
	g.p("ctrl: s.ctrl,")
	g.p("txOpened: true,")
	g.p("}")
	g.br()
	g.p("return newClient, nil")
	g.p("}")
}

func (g *generator) GenerateTx(dstPkg looker.ImportElement, intf *looker.Interface) {
	g.p("func (s *%s) Tx() sal.Transaction {", intf.ImplementationName(Prefix))
	g.p("if tx, ok := s.handler.(sal.SqlTx); ok {")
	g.p("return sal.NewWrappedTx(tx, s.ctrl)")
	g.p("}")
	g.p("return nil")
	g.p("}")
}

func (g *generator) ifErr(resp, msg string) {
	g.p("if err != nil {")
	if resp == "" {
		g.p("return errors.Wrap(err, %q)", msg)
	} else {
		g.p("return %s, errors.Wrap(err, %q)", resp, msg)
	}
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
	if isSqlResult(prms[0]) {
		return sal.OperationTypeExec
	}
	if prms[0].Kind() == reflect.Slice.String() {
		return sal.OperationTypeQuery
	}
	return sal.OperationTypeQueryRow
}

func ImportPaths(dirtyList []string, dstPath string) []string {
	list := make([]string, 0)
	for _, p := range dirtyList {
		// todo: find mistake when import contains something from vendor
		if strings.Contains(p, "/vendor/") {
			continue
		}
		if p != "" && p != dstPath {
			list = append(list, p)
		}
	}

	return list
}

func responseErrStr(operation sal.OperationType, resp looker.Parameter, dstPath string) string {
	var errRespStr string
	if operation != sal.OperationTypeExec {
		if resp.Pointer() {
			errRespStr = "nil"
		} else {
			if resp.Kind() == reflect.Struct.String() {
				errRespStr = resp.Name(dstPath) + "{}"
			} else {
				errRespStr = "nil"
			}
		}
	}

	return errRespStr
}

func isSqlResult(prm looker.Parameter) bool {
	if prm.Name("path/to/pkg") == "sql.Result" {
		return true
	}
	return false
}
