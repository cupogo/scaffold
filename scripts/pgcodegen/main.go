// go:build codegen
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
	"hyyl.xyz/cupola/scaffold/pkg/utils"
)

var (
	dropfirst bool
	doc       *Document
)

func init() {
	flag.BoolVar(&dropfirst, "drop", false, "drop exists first")
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Print("usage: codegen filename")
		return
	}
	docfile := args[0]

	var err error
	doc, err = NewDoc(docfile)
	if err != nil {
		log.Printf("load fail: %s", err)
		return
	}

	if err = doc.genModels(dropfirst); err == nil {
		err = doc.genStores(dropfirst)
	}

	if err != nil {
		log.Printf("output fail: %s", err)
		return
	} else {
		log.Print("generated ok")
	}
}

func getQual(k string) (string, bool) {
	if doc != nil {
		return doc.getQual(k)
	}
	return "", false
}

func getTableName(mn string) string {
	if doc != nil {
		for _, model := range doc.Models {
			if model.Name == mn {
				return model.tableName()
			}
		}
	}
	return utils.Underscore(mn)
}

func validModel(name string) bool {
	if doc != nil {
		for _, model := range doc.Models {
			if model.Name == name {
				return true
			}
		}
	}

	return false
}

func loadPackage(path string) *packages.Package {
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedImports}
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
