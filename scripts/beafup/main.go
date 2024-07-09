package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/cupogo/scaffold/templates"
)

var (
	name string
)

func main() {
	flag.StringVar(&name, "name", "", "prefix of model")
	flag.Parse()

	if len(name) == 0 {
		flag.Usage()
		return
	}
	name = strings.ToLower(name)

	src := "docs/demo.yaml"
	dst := "docs/" + name + ".yaml"

	_, err := os.Stat(dst)
	if err == nil || !os.IsNotExist(err) {
		log.Printf("already exist: %s", dst)
		return
	}

	_ = templates.Render(src, dst, &Data{
		Name: name,
	})
}

type Data struct {
	Name string
}
