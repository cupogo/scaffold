// go:build codegen
package main

import (
	"flag"
	"log"

	"github.com/cupogo/scaffold/scripts/codegen/gens"
)

var (
	dropfirst bool
	doc       *gens.Document
	genSpec   int
)

func init() {
	dftSpec := int(gens.TgModel + gens.TgStore + gens.TgWeb)
	flag.BoolVar(&dropfirst, "drop", false, "drop exists first")
	flag.IntVar(&genSpec, "spec", dftSpec, "which spec to generate")
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

	gens.Run(docfile, gens.TagSpec(genSpec), dropfirst)
}
