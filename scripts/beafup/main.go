package main

import (
	"flag"
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

	templates.Render(src, dst, &Data{
		Name: name,
	})
}

type Data struct {
	Name string
}
