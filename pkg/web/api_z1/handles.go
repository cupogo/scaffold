package apiz1

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/cupogo/scaffold/scripts/codegen/gens"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
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
	regHI(false, "POST", "/docs/yamls/:name", "", func(a *api) gin.HandlerFunc {
		return a.postDocsYaml
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

func (a *api) postDocsYaml(c *gin.Context) {
	name := c.Param("name")
	if !strings.HasSuffix(name, ".yaml") {
		name = name + ".yaml"
	}

	logger().Infow("posted", "name", name, "ct", c.ContentType(), "cl", c.Request.ContentLength)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fail(c, 400, err)
		return
	}
	if err = c.Request.Body.Close(); err != nil {
		logger().Infow("close body fail", "err", err)
	}
	var doc gens.Document
	err = yaml.Unmarshal(data, &doc)
	if err != nil {
		fail(c, 400, err)
		return
	}
	if err = doc.Check(); err != nil {
		fail(c, 400, err)
		return
	}

	err = os.WriteFile(path.Join(Root, "docs", name), data, 0644)
	if err != nil {
		fail(c, 500, err)
		return
	}
	success(c, "ok")
}
