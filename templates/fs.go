package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
)

//go:embed */*.tmpl
var tplfs embed.FS

func FS() fs.FS {
	return &tplfs
}

func Render(src, dest string, data any) error {
	t := template.Must(template.ParseFS(tplfs, src+".tmpl"))
	wr, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = t.Execute(wr, data)
	if err != nil {
		slog.Info("render fail", "src", src, "err", err)
		os.Remove(dest)
	}
	return err
}
