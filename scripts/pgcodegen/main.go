// go:build codegen
package main

import (
	"flag"
	"log"

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
	doc.prepare()

	if genSpec&tgModel > 0 {
		if err = doc.genModels(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		// log.Print("generated models ok")
	}
	if genSpec&tgStore > 0 {
		if err = doc.genStores(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		// log.Print("generated stores ok")
	}
	if genSpec&tgWeb > 0 {
		if err = doc.genWebAPI(); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		// log.Print("generated webapi ok")
	}

}

func getQual(k string) (string, bool) {
	if doc != nil {
		return doc.getQual(k)
	}
	return "", false
}

func getModQual(k string) (string, bool) {
	if doc != nil {
		return doc.getModQual(k)
	}
	return k, false
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
