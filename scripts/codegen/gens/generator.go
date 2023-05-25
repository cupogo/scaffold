package gens

import (
	"log"
)

type TagSpec uint8

const (
	TgModel TagSpec = 1 << iota
	TgStore
	TgWeb
)

var curgen *Generator

func init() {
	curgen = new(Generator)
}

type Generator struct {
	doc *Document
}

func Run(docfile string, genSpec TagSpec, dropfirst bool) {

	var err error
	curgen.doc, err = NewDoc(docfile)
	if err != nil {
		log.Printf("load fail: %s", err)
		return
	}

	err = curgen.doc.Check()
	if err != nil {
		log.Printf("check fail: %s", err)
		return
	}
	curgen.doc.Init()

	if genSpec&TgModel > 0 {
		if err = curgen.doc.genModels(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		// log.Print("generated models ok")
	}
	if genSpec&TgStore > 0 {
		if err = curgen.doc.genStores(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		// log.Print("generated stores ok")
	}
	if genSpec&TgWeb > 0 {
		if err = curgen.doc.genWebAPI(dropfirst); err != nil {
			log.Printf("output fail: %s", err)
			return
		}
		// log.Print("generated webapi ok")
	}
}

func getQual(k string) (string, bool) {
	if curgen.doc != nil {
		return curgen.doc.getQual(k)
	}
	return "", false
}
