package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/utils/array"
)

type Permission struct {
	comm.DunceModel

	Path     string
	Method   string
	Name     string
	IsActive bool
}

type Permissions []Permission

func main() {

	data, err := genPermissions("./docs/swagger.yaml")
	if err != nil {
		return
	}
	dir := "./database/schemas"
	if !IsDir(dir) {
		slog.Warn("directory not found", "dir", dir)
		return
	}

	genPermissionSql(data, filepath.Join(dir, "pg_10_auth_permissions.sql"))
}

func genPermissions(fdoc string) (data Permissions, err error) {
	doc, err := loadDoc(fdoc)
	if err != nil {
		slog.Error("failed", "err", err)
		return
	}
	for path, methods := range doc.Paths {
		for method, entry := range methods {
			if entry.OperationID == "" { // 跳过无ID的
				continue
			}
			var obj = new(Permission)
			obj.Path = path
			obj.Method = strings.ToUpper(method)
			obj.ID = entry.OperationID
			obj.Name = "API: " + entry.Summary
			obj.IsActive = true
			data = append(data, *obj)
		}
	}

	sort.Slice(data, func(i, j int) bool { return data[i].ID < data[j].ID })
	return
}

func genPermissionSql(perms Permissions, path string) error {
	df, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = df.Truncate(0)
		}
		df.Close()
	}()

	prefixes := array.NewString()

	// init updated
	_, err = fmt.Fprintf(df,
		`-- auth permission data
INSERT INTO auth_permission ("id", "creator_id", "updated", "name", "path", "method", "is_active", "remark") VALUES %c`, '\n')
	if err != nil {
		return err
	}
	for i, p := range perms {
		if a, _, ok := strings.Cut(p.ID, "-"); ok && len(a) > 0 {
			prefixes.Insert(a)
		}
		_, err = fmt.Fprintf(df, `('%s', %d, CURRENT_TIMESTAMP, '%s', '%s', '%s', %s, '')`,
			p.ID, p.CreatorID, p.Name, p.Path, p.Method, strconv.FormatBool(p.IsActive))
		if err != nil {
			return err
		}
		if i == len(perms)-1 {
			_, err = fmt.Fprint(df, "\nON CONFLICT (id) DO UPDATE SET updated = CURRENT_TIMESTAMP;\n\n")
			if err != nil {
				return err
			}
		} else {
			_, err = fmt.Fprint(df, ",\n")
			if err != nil {
				return err
			}
		}

	}

	pres := strings.Join(prefixes.List(), "|")
	if len(prefixes) > 1 {
		pres = fmt.Sprintf("(%s)", pres)
	}

	_, err = fmt.Fprintf(df, "DELETE FROM auth_permission WHERE id SIMILAR TO '%s-%%' AND updated < CURRENT_DATE -1;\n",
		pres)
	if err != nil {
		return err
	}

	slog.Info("gen sql file ok", "to", path, "count", len(perms), "pres", pres)
	return err
}

// IsDir ...
func IsDir(fpath string) bool {
	fi, err := os.Stat(fpath)
	return err == nil && fi.Mode().IsDir()
}
