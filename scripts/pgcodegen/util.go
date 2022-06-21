// go:build codegen
package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
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

func checkPackageObject(pkg *packages.Package, name string) bool {
	for _, f := range pkg.Syntax {
		// log.Printf("Decls %+v", f.Decls)
		// log.Printf("Scope %+v", f.Scope)
		for k := range f.Scope.Objects {
			// log.Printf("object %s: %v", k, obj)
			if k == name {
				return true
			}
		}
	}
	return false
}

type Cursor = astutil.Cursor

func loadTypes(name string) (objs []ast.Object) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, nil, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		log.Printf("parse %s fail %s", name, err)
		return
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.StructType); ok {
				objs = append(objs, *x.Name.Obj)
			}
			if _, ok := x.Type.(*ast.ArrayType); ok {
				objs = append(objs, *x.Name.Obj)
			}
		}

		return true
	})

	return
}

type vast struct {
	name string
	fset *token.FileSet
	file *ast.File

	body []byte
}

func newAST(name string) (*vast, error) {
	fset := token.NewFileSet()
	var err error
	file, err := parser.ParseFile(fset, name, nil, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		log.Printf("parse %s fail %s", name, err)
		return nil, err
	}
	o := &vast{
		name: name,
		fset: fset,
		file: file,
	}

	return o, nil
}

func (w *vast) rewrite(pre, post astutil.ApplyFunc) (ok bool) {
	n := astutil.Apply(w.file, pre, post)
	var buf bytes.Buffer
	if err := format.Node(&buf, w.fset, n); err != nil {
		log.Printf("format fail %s", err)
		return
	}
	w.body = buf.Bytes()
	ok = true
	return
}

func (w *vast) addStructField(fields *ast.FieldList, name, typ string) {
	// prevField := fields.List[fields.NumFields()-1]

	c := &ast.Comment{Text: "// " + name + " gened" /*Slash: prevField.End() + 1*/}
	cg := &ast.CommentGroup{List: []*ast.Comment{c}}
	o := ast.NewObj(ast.Var, name)
	f := &ast.Field{
		Comment: cg,
		Names:   []*ast.Ident{{Name: name, Obj: o, NamePos: cg.End() + 1}},
	}
	o.Decl = f
	f.Type = &ast.StarExpr{X: &ast.Ident{Name: typ, NamePos: f.Names[0].End() + 1}}

	w.fset.File(c.End()).AddLine(int(c.End()))
	w.fset.File(f.End()).AddLine(int(f.End()))

	fields.List = append(fields.List, f)
	// w.file.Comments = append(w.file.Comments, cg)
}

func fieldecl(name, typ string) *ast.Field {
	return &ast.Field{
		Names:   []*ast.Ident{ast.NewIdent(name)},
		Type:    &ast.StarExpr{X: ast.NewIdent(typ)},
		Comment: &ast.CommentGroup{List: []*ast.Comment{{Text: "// with gen"}}},
	}
}

func existVarField(list *ast.FieldList, name string) bool {
	for _, field := range list.List {
		for _, id := range field.Names {
			if id.Obj.Kind == ast.Var && id.Name == name {
				log.Printf("exist field %s", name)
				return true
			}
		}

	}

	return false
}

func valspec(name, typ string) *ast.ValueSpec {
	return &ast.ValueSpec{Names: []*ast.Ident{ast.NewIdent(name)},
		Type: ast.NewIdent(typ),
	}
}

func vardecl(name, typ string) *ast.GenDecl {
	return &ast.GenDecl{
		Tok:   token.VAR,
		Specs: []ast.Spec{valspec(name, typ)},
	}
}

func showNode(n ast.Node) []byte {
	var buf bytes.Buffer
	fset := token.NewFileSet()

	if err := format.Node(&buf, fset, n); err != nil {
		log.Printf("show node fail %s", err)
		return nil
	}
	return buf.Bytes()
}

func wnasstmt(name string) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("w"), Sel: ast.NewIdent(name)}},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{&ast.UnaryExpr{Op: token.AND, X: &ast.CompositeLit{
			Type: ast.NewIdent(name),
			Elts: []ast.Expr{ast.NewIdent("w")},
		}}},
	}
}

func wrapNewFunc(s *Store, prev ast.Node) *ast.FuncDecl {
	siname := s.ShortIName()
	c := &ast.Comment{Text: "// " + siname + " gened", Slash: prev.End() + 1}
	cg := &ast.CommentGroup{List: []*ast.Comment{c}}
	return &ast.FuncDecl{
		Doc: cg,
		Recv: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("w")},
			Type:  &ast.StarExpr{X: ast.NewIdent(storewn)},
		}}},
		Name: ast.NewIdent(siname),
		Type: &ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{
			{Type: ast.NewIdent(s.IName)},
		}}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
			&ast.SelectorExpr{X: ast.NewIdent("w"), Sel: ast.NewIdent(s.Name)},
		}}}},
	}
}

func existBlockAssign(block *ast.BlockStmt, name string) bool {
	for _, st := range block.List {
		if as, ok := st.(*ast.AssignStmt); ok {
			if len(as.Lhs) > 0 && len(as.Rhs) > 0 {
				if se, ok := as.Lhs[0].(*ast.SelectorExpr); ok && se.Sel.Name == name {
					log.Printf("exist assign %s", name)
					return true
				}
			}
		}
	}
	return false
}
