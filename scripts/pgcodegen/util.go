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

func rewriteAST(name string, pre, post astutil.ApplyFunc) ([]byte, bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, nil, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		log.Printf("parse %s fail %s", name, err)
		return nil, false
	}

	n := astutil.Apply(file, pre, post)
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, n); err != nil {
		log.Printf("format %s fail %s", name, err)
		return nil, false
	}
	return buf.Bytes(), true
}
