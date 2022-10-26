// go:build codegen
package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/jinzhu/inflection"
	"golang.org/x/tools/go/packages"

	"github.com/cupogo/scaffold/templates"
)

type empty struct{}

func loadPackage(path string) *packages.Package {
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedCompiledGoFiles |
		packages.NeedTypes | packages.NeedSyntax | packages.NeedImports}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatalf("loading packages for inspection: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0]
}

func fieldecl(name, typ string) *dst.Field {
	f := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  &dst.StarExpr{X: dst.NewIdent(typ)},
	}
	f.Decorations().End.Append("// gened")

	return f
}

func existVarField(list *dst.FieldList, name string) bool {
	for _, field := range list.List {
		for _, id := range field.Names {
			if id.Obj.Kind == dst.Var && id.Name == name {
				// log.Printf("exist field %s", name)
				return true
			}
		}

	}

	return false
}

// nolint
func showNode(node dst.Node) string {
	before, after, points := dstutil.Decorations(node)
	var info string
	if before != dst.None {
		info += fmt.Sprintf("- Before: %s\n", before)
	}
	for _, point := range points {
		if len(point.Decs) == 0 {
			continue
		}
		info += fmt.Sprintf("- %s: [", point.Name)
		for i, dec := range point.Decs {
			if i > 0 {
				info += ", "
			}
			info += fmt.Sprintf("%q", dec)
		}
		info += "]\n"
	}
	if after != dst.None {
		info += fmt.Sprintf("- After: %s\n", after)
	}
	if info != "" {
		fmt.Printf("%T\n%s\n", node, info)
	}
	return info
}

func wnasstmt(name string) *dst.AssignStmt {
	st := &dst.AssignStmt{
		Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent(name)}},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{
			Type: dst.NewIdent(name),
			Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("w"), Value: dst.NewIdent("w")}},
		}}},
	}
	st.Decs.Before = dst.None
	st.Decs.After = dst.None
	st.Decorations().End.Append("// gened")
	return st
}

func shimNode(n dst.Node) {
	if n.Decorations().After == dst.EmptyLine && len(n.Decorations().End) > 0 {
		n.Decorations().After = dst.NewLine
	}
	if n.Decorations().Before == dst.EmptyLine {
		n.Decorations().Before = dst.NewLine
	}
}

func wrapNewFunc(s *Store, prev dst.Node) *dst.FuncDecl {
	siname := s.ShortIName()
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{List: []*dst.Field{{
			Names: []*dst.Ident{dst.NewIdent("w")},
			Type:  &dst.StarExpr{X: dst.NewIdent(storewn)},
		}}},
		Name: dst.NewIdent(siname),
		Type: &dst.FuncType{Results: &dst.FieldList{List: []*dst.Field{
			{Type: dst.NewIdent(s.GetIName())},
		}}},
		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{
			&dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent(s.Name)},
		}}}},
	}
	// f.Decorations().Start.Prepend("\n")
	f.Decorations().End.Append("// " + siname + " gened")

	return f
}

func existBlockAssign(block *dst.BlockStmt, name string) bool {
	for _, st := range block.List {
		if as, ok := st.(*dst.AssignStmt); ok {
			if len(as.Lhs) > 0 && len(as.Rhs) > 0 {
				if se, ok := as.Lhs[0].(*dst.SelectorExpr); ok && se.Sel.Name == name {
					// log.Printf("exist assign %s", name)
					return true
				}
			}
		}
	}
	return false
}

type vdst struct {
	name string
	fset *token.FileSet
	file *dst.File

	body []byte
}

func newDST(name string) (*vdst, error) {
	fset := token.NewFileSet()
	var err error
	file, err := decorator.ParseFile(fset, name, nil, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		log.Printf("parse %s fail %s", name, err)
		return nil, err
	}

	o := &vdst{
		name: name,
		fset: fset,
		file: file,
	}

	return o, nil

}

func (w *vdst) rewrite(pre, post dstutil.ApplyFunc) (ok bool) {
	n := dstutil.Apply(w.file, pre, post).(*dst.File)
	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, n); err != nil {
		log.Printf("format fail %s", err)
		return
	}
	w.body = buf.Bytes()
	ok = true
	return
}

func existInterfaceMethod(it *dst.InterfaceType, name string) bool {
	for _, field := range it.Methods.List {
		for _, id := range field.Names {
			if id.Name == name {
				return true
			}
		}
	}
	return false
}

func newStoInterfaceMethod(name, ret string) *dst.Field {
	id := dst.NewIdent(name)
	id.Obj = dst.NewObj(dst.Fun, name)
	f := &dst.Field{
		Names: []*dst.Ident{id},
		Type: &dst.FuncType{
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: dst.NewIdent(ret)},
				},
			},
		},
	}
	f.Decorations().End.Append("// gened")

	return f
}

func cutMethod(s string) (act string, tgt string, ok bool) {
	var foundLow bool
	var foundUp bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsUpper(c) {
			if foundUp && foundLow || i > 2 { // PutObject, putObject
				act = s[0:i]
				tgt = s[i:]
				ok = len(tgt) > 0
				return
			}
			foundUp = true
		}
		if IsLower(c) {
			foundLow = true
		}
	}

	return
}

// CheckFile returns true if a file exists
func CheckFile(fpath string) (exists bool) {
	_, err := os.Stat(fpath)
	exists = !os.IsNotExist(err)
	return
}

// IsDir ...
func IsDir(fpath string) bool {
	fi, err := os.Stat(fpath)
	return err == nil && fi.Mode().IsDir()
}

func Plural(str string) string {
	if strings.HasSuffix(str, "ID") {
		return str + "s"
	}
	return inflection.Plural(str)
}

func HookMethod(model string, k, v string) (string, bool) {
	if strings.HasPrefix(v, "db") {
		return v, true
	}
	if v == "afterCreate"+model { // deprecated
		return v, true
	}
	if v == "true" || v == "yes" { // true, yes
		if strings.HasSuffix(k, "ing") {
			return "db" + ToExported(k[0:len(k)-3]+"e") + model, true
		}
	}
	switch k {
	case afterLoad, afterList, beforeList, afterCreated, afterUpdated, afterDeleted:
		return k + model, true
	}
	return "", false
}

func ensureGoFile(gfile, tplname string, data any) {
	if !CheckFile(gfile) {
		if err := renderTmpl(tplname, gfile, data); err != nil {
			panic(err)
		}
		log.Printf("write go file ok, %s", gfile)
	}
}

func renderTmpl(src, dest string, data any) error {
	tplf := src + ".go.tmpl"
	t := template.Must(template.ParseFS(templates.FS(), tplf))
	wr, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = t.Execute(wr, data)
	if err != nil {
		log.Printf("render fail: %s", err)
		os.Remove(dest)
	}
	return err
}
