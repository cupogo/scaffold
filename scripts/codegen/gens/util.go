// go:build codegen
package gens

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/guess"
	"github.com/dave/dst/dstutil"
	"github.com/dave/jennifer/jen"
	"github.com/jinzhu/inflection"
	"golang.org/x/tools/go/packages"

	"github.com/cupogo/scaffold/pkg/services/utils"
	"github.com/cupogo/scaffold/templates"
)

type empty struct{}

func GetModule(dir string) (string, error) {
	return utils.GetModule(dir)
}

func loadPackage(path string) *packages.Package {
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}

	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedTypes | packages.NeedSyntax | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatalf("loading packages for inspection: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0]
}

func isFieldInList(list *dst.FieldList, name string) bool {
	for _, field := range list.List {
		for _, id := range field.Names {
			if id.Name == name {
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

func shimNode(n dst.Node) {
	if n.Decorations().After == dst.EmptyLine && len(n.Decorations().End) > 0 {
		n.Decorations().After = dst.NewLine
	}
	if n.Decorations().Before == dst.EmptyLine {
		n.Decorations().Before = dst.NewLine
	}
}

// func wrapNewFunc(s *Store, prev dst.Node) *dst.FuncDecl {
// 	siname := s.ShortIName()
// 	f := &dst.FuncDecl{
// 		Recv: &dst.FieldList{List: []*dst.Field{newField("w", storewn, true)}},
// 		Name: dst.NewIdent(siname),
// 		Type: &dst.FuncType{Results: &dst.FieldList{List: []*dst.Field{
// 			{Type: dst.NewIdent(s.GetIName())},
// 		}}},
// 		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{
// 			&dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent(s.Name)},
// 		}}}},
// 	}
// 	// f.Decorations().Start.Prepend("\n")
// 	f.Decorations().End.Append("// " + siname + " gened")

// 	return f
// }

func existBlockAssign(block *dst.BlockStmt, name string) (int, bool) {
	for i, st := range block.List {
		if as, ok := st.(*dst.AssignStmt); ok {
			if len(as.Lhs) > 0 && len(as.Rhs) > 0 {
				if se, ok := as.Lhs[0].(*dst.SelectorExpr); ok && se.Sel.Name == name {
					// log.Printf("exist assign %s", name)
					return i, true
				}
			}
		}
	}
	return -1, false
}

type vdst struct {
	name string
	pkgn string
	fset *token.FileSet
	file *dst.File
}

func newDST(name, pkg string) (*vdst, error) {
	fset := token.NewFileSet()
	dec := decorator.NewDecoratorWithImports(fset, "main", goast.New())
	var err error
	file, err := dec.ParseFile(name, nil, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		log.Printf("parse %s fail %s", name, err)
		return nil, err
	}

	o := &vdst{
		name: name,
		pkgn: pkg,
		fset: fset,
		file: file,
	}

	return o, nil

}

func (v *vdst) existFunc(name string) bool {
	for _, decl := range v.file.Decls {
		if fd, ok := decl.(*dst.FuncDecl); ok && fd.Name.Name == name {
			return true
		}
	}
	return false
}

func (v *vdst) ensureFunc(name string, fd *dst.FuncDecl) {
	if fd != nil && !v.existFunc(name) {
		v.file.Decls = append(v.file.Decls, fd)
		log.Printf("ensureFunc: %s", name)
	}
}

func (v *vdst) existType(name string) bool {
	for _, decl := range v.file.Decls {
		if gd, ok := decl.(*dst.GenDecl); ok && gd.Tok == token.TYPE {
			for _, sp := range gd.Specs {
				if tsp, ok := sp.(*dst.TypeSpec); ok && tsp.Name.Name == name {
					return true
				}
			}
		}
	}
	return false
}

// nolint
func (v *vdst) existMethod(name, recv string) (int, bool) {
	for idx, decl := range v.file.Decls {
		if fd, ok := decl.(*dst.FuncDecl); ok && fd.Name.Name == name {
			if fd.Recv != nil && len(fd.Recv.List) == 1 {
				if ex, ok := fd.Recv.List[0].Type.(*dst.Ident); ok && ex.Name == recv {
					return idx, true
				}
				if se, ok := fd.Recv.List[0].Type.(*dst.StarExpr); ok {
					if ex, ok := se.X.(*dst.Ident); ok && ex.Name == recv {
						return idx, true
					}
				}
			}
		}
	}
	return -1, false
}

// nolint
func (v *vdst) ensureMethod(name, recv string, fd *dst.FuncDecl) {
	if fd == nil {
		return
	}
	if idx, ok := v.existMethod(name, recv); ok {
		v.file.Decls[idx] = fd
	} else {
		v.file.Decls = append(v.file.Decls, fd)
	}

	log.Printf("ensureFunc: %s", name)
}

func (v *vdst) Apply(pre, post dstutil.ApplyFunc) dst.Node {
	return dstutil.Apply(v.file, pre, post)
}

func (w *vdst) overwrite() error {
	res := decorator.NewRestorerWithImports(w.pkgn, guess.New())
	var buf bytes.Buffer
	if err := res.Fprint(&buf, w.file); err != nil {
		log.Printf("format fail: %s", err)
		return err
	}
	if err := os.WriteFile(w.name, buf.Bytes(), 0644); err != nil {
		log.Printf("write file %q fail: %s", w.name, err)
		return err
	}
	log.Printf("write file %q ok", w.name)
	return nil
}

// newField return dst.Field with name, type, is pointer
func newField(vn any, vt any, star bool) *dst.Field {
	var id *dst.Ident
	if name, ok := vn.(string); ok {
		id = dst.NewIdent(name)
	} else if n, ok := vn.(*dst.Ident); ok {
		id = n
	} else {
		panic(fmt.Errorf("invalid vn: %+v", vn))
	}
	f := &dst.Field{Names: []*dst.Ident{id}}

	if id, ok := vt.(*dst.Ident); ok {
		if star {
			f.Type = &dst.StarExpr{X: id}
			return f
		}
		f.Type = id
		return f
	}

	if ex, ok := vt.(dst.Expr); ok {
		f.Type = ex
		return f
	}

	if s, ok := vt.(string); ok {
		if star {
			f.Type = &dst.StarExpr{X: dst.NewIdent(s)}
			return f
		}
		f.Type = dst.NewIdent(s)
		return f
	}

	panic(fmt.Errorf("invalid vt: %+v", vt))
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

func ParseHook(model, ns string, k, v string) (sh storeHook, ok bool) {
	sh.k = k
	sh.ObjName = model
	if strings.HasPrefix(v, "db") {
		sh.FunName = v
		return sh, true
	}
	if v == "afterCreate"+model { // deprecated
		sh.FunName = v
		return sh, true
	}
	suf := model
	if len(ns) > 0 {
		suf = ToExported(ns) + model
	}
	var deco string
	v, deco, _ = strings.Cut(v, ",")
	sh.isPtr = strings.Contains(deco, "ptr")
	sh.isTot = strings.Contains(deco, "tot")
	if v == "true" || v == "yes" { // true, yes
		if strings.HasSuffix(k, "ing") {
			sh.FunName = "db" + ToExported(k[0:len(k)-3]+"e") + suf
			return sh, true
		}
		v = k + suf
	}
	switch k {
	case afterLoad, beforeList, afterCreated, afterUpdated, afterDeleted, upsertES, deleteES:
		sh.FunName = v
		return sh, strings.HasPrefix(v, k+suf)
	case afterList:
		sh.FunName = v
		return sh, strings.HasPrefix(v, k+suf)
	case errorLoad:
		sh.FunName = "on" + ToExported(k) + suf
		return sh, true
	}
	return
}

func ensureGoFile(gfile, tplname string, data any) {
	if !CheckFile(gfile) {
		if err := templates.Render(tplname+".go", gfile, data); err != nil {
			panic(err)
		}
		log.Printf("write go file ok, %s", gfile)
	}
}

func matchs(patt string, names ...string) bool {
	for _, name := range names {
		if ok, _ := path.Match(patt, name); ok {
			return true
		}
	}

	return false
}

// nolint
func dstExpr(expr string) (dst.Node, error) {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return nil, err
	}
	dec := decorator.NewDecorator(nil)
	return dec.DecorateNode(node)
}

// nolint
func pickExpr(expr string) (out string, err error) {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return
	}

	fset := token.NewFileSet()

	var buf bytes.Buffer
	err = format.Node(&buf, fset, node)
	if err != nil {
		return
	}

	out = buf.String()

	return
}

func jcodeDesc(st *jen.Statement, txt, pre string) {
	if len(txt) > 0 {
		for _, line := range strings.Split(txt, "\n") {
			if len(line) > 0 {
				st.Comment(pre + line).Line()
			}
		}
	}
}

func highlights(a []string) []string {
	o := make([]string, len(a))
	for i := range a {
		o[i] = "`" + a[i] + "`"
	}
	return o
}
