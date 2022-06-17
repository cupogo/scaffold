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

const (
	tgModel = 1 << iota
	tgStore
	tgWeb
)

var (
	dropfirst bool
	doc       *Document
	genSpec   int
)

func init() {
	flag.BoolVar(&dropfirst, "drop", false, "drop exists first")
	flag.IntVar(&genSpec, "spec", tgModel+tgStore+tgWeb, "which spec to generate")
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

	if genSpec&tgModel > 0 {
		if err = doc.genModels(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		log.Print("generated models ok")
	}
	if genSpec&tgStore > 0 {
		if err = doc.genStores(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		log.Print("generated stores ok")
	}
	if genSpec&tgWeb > 0 {
		if err = doc.genWebAPI(); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		log.Print("generated webapi ok")
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
		if model, ok := doc.modelWithName(mn); ok {
			return model.tableName()
		}
	}
	return utils.Underscore(mn)
}

func getModel(name string) *Model {
	if doc != nil {
		if model, ok := doc.modelWithName(name); ok {
			return model
		}
	}

	return &Model{}
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

func cutMethod(s string) (act string, tgt string, ok bool) {
	var foundLow bool
	var foundUp bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		if utils.IsUpper(c) {
			if foundUp && foundLow || i > 2 { // PutObject, putObject
				act = s[0:i]
				tgt = s[i:]
				ok = len(tgt) > 0
				return
			}
			foundUp = true
		}
		if utils.IsLower(c) {
			foundLow = true
		}
	}

	return
}
