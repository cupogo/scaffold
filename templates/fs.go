// go:build codegen
package templates

import (
	"embed"
	"io/fs"
)

//go:embed */*.tmpl
var tplfs embed.FS

func FS() fs.FS {
	return &tplfs
}
