// go:build codegen
package main

import (
	"flag"
	"log"
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
	doc.Init()

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
