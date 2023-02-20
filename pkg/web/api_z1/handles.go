package apiz1

import (
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	Root = "./"
)

func init() {
	regHI(false, "GET", "/docs/yamls", "", func(a *api) gin.HandlerFunc {
		return a.getDocsYamls
	})
	regHI(false, "GET", "/docs/yamls/:name", "", func(a *api) gin.HandlerFunc {
		return a.getDocsYaml
	})

}

type DocEntry struct {
	Name  string      `json:"name"`
	Mode  os.FileMode `json:"mode,omitempty"`
	IsDir bool        `json:"isDir,omitempty"`
	Size  int64       `json:"size,omitempty"`
}

func (a *api) getDocsYamls(c *gin.Context) {
	files, err := os.ReadDir(path.Join(Root, "docs"))
	if err != nil {
		fail(c, 400, err)
		return
	}

	var data []DocEntry
	for _, f := range files {
		var size int64
		if fi, err := f.Info(); err == nil {
			size = fi.Size()
		}
		name := f.Name()
		if strings.HasPrefix(name, "swagger.") {
			continue
		}
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		data = append(data, DocEntry{
			Name:  name,
			Mode:  f.Type(),
			IsDir: f.IsDir(),
			Size:  size,
		})
	}

	success(c, data)
}

func (a *api) getDocsYaml(c *gin.Context) {
	name := c.Param("name")
	if !strings.HasSuffix(name, ".yaml") {
		name = name + ".yaml"
	}
	data, err := os.ReadFile(path.Join(Root, "docs", name))
	if err != nil {
		fail(c, 400, err)
		return
	}
	c.Data(200, "text/yaml", data)
}
