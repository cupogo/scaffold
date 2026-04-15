package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/models/oid"
	"github.com/cupogo/andvari/utils/array"
)

type ApiRecord struct {
	comm.DefaultModel

	OperationID string
	Endpoint    string
	Method      string
	Summary     string
	Description string
	Parameters  string // JSON string
	Responses   string // JSON string
}

type ApiRecords []ApiRecord

type Permission struct {
	comm.DunceModel

	Path     string
	Method   string
	Name     string
	IsActive bool
}

type Permissions []Permission

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: go run ./scripts/sqlgen [perm|api]")
		os.Exit(1)
	}
	cmd := os.Args[1]

	dir := "./database/schemas"
	if !IsDir(dir) {
		slog.Warn("directory not found", "dir", dir)
		os.Exit(1)
	}

	switch cmd {
	case "perm":
		data, err := genPermissions("./docs/swagger.yaml")
		if err != nil {
			os.Exit(1)
		}
		err = genPermissionSql(data, filepath.Join(dir, "pg_10_auth_permissions.sql"))
		if err != nil {
			os.Exit(1)
		}
	case "api":
		data, err := genApiRecords("./docs/swagger.yaml")
		if err != nil {
			os.Exit(1)
		}
		err = genApiRecordSql(data, filepath.Join(dir, "pg_11_api_records.sql"))
		if err != nil {
			os.Exit(1)
		}
	default:
		slog.Error("unknown command", "cmd", cmd)
		os.Exit(1)
	}
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

func genApiRecords(fdoc string) (data ApiRecords, err error) {
	doc, err := loadDoc(fdoc)
	if err != nil {
		slog.Error("failed to load doc", "err", err)
		return
	}
	for path, methods := range doc.Paths {
		for method, entry := range methods {
			var obj = new(ApiRecord)
			obj.OperationID = entry.OperationID
			obj.Endpoint = path
			obj.Method = strings.ToUpper(method)
			obj.Summary = entry.Summary
			if entry.Description != "" {
				obj.Description = entry.Description
			} else {
				obj.Description = entry.Summary
			}

			// Serialize parameters to JSON
			if len(entry.Parameters) > 0 {
				paramsJSON, jerr := json.Marshal(entry.Parameters)
				if jerr != nil {
					slog.Warn("failed to marshal parameters", "err", jerr, "path", path, "method", method)
					obj.Parameters = "[]"
				} else {
					obj.Parameters = string(paramsJSON)
				}
			} else {
				obj.Parameters = "[]"
			}

			// Serialize responses to JSON
			if len(entry.Responses) > 0 {
				respJSON, jerr := json.Marshal(entry.Responses)
				if jerr != nil {
					slog.Warn("failed to marshal responses", "err", jerr, "path", path, "method", method)
					obj.Responses = "{}"
				} else {
					obj.Responses = string(respJSON)
				}
			} else {
				obj.Responses = "{}"
			}

			obj.CreatorID = 0
			obj.ID = oid.NewID(oid.OtDefault)
			data = append(data, *obj)
		}
	}

	sort.Slice(data, func(i, j int) bool {
		if data[i].Endpoint != data[j].Endpoint {
			return data[i].Endpoint < data[j].Endpoint
		}
		return data[i].Method < data[j].Method
	})
	return
}

func genApiRecordSql(records ApiRecords, path string) error {
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

	_, err = fmt.Fprintf(df,
		`-- api_record data
INSERT INTO api_record ("id", "operation_id", "endpoint", "method", "summary", "description", "parameters", "responses", "creator_id", "created", "updated") VALUES %c`, '\n')
	if err != nil {
		return err
	}

	for i, r := range records {
		// Escape single quotes in strings
		summary := strings.ReplaceAll(r.Summary, "'", "''")
		description := strings.ReplaceAll(r.Description, "'", "''")

		_, err = fmt.Fprintf(df, `(%d, '%s', '%s', '%s', '%s', '%s', '%s'::jsonb, '%s'::jsonb, %d, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			r.ID, r.OperationID, r.Endpoint, r.Method, summary, description, r.Parameters, r.Responses, r.CreatorID)
		if err != nil {
			return err
		}
		if i == len(records)-1 {
			_, err = fmt.Fprint(df, "\nON CONFLICT (endpoint, method) DO UPDATE SET\n    operation_id = EXCLUDED.operation_id,\n    summary = EXCLUDED.summary,\n    description = EXCLUDED.description,\n    parameters = EXCLUDED.parameters,\n    responses = EXCLUDED.responses,\n    updated = CURRENT_TIMESTAMP;\n\n")
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

	_, err = fmt.Fprintf(df, "DELETE FROM api_record WHERE updated < CURRENT_DATE - INTERVAL '1 day';\n")
	if err != nil {
		return err
	}

	slog.Info("gen api_record sql ok", "to", path, "count", len(records))
	return nil
}

// IsDir ...
func IsDir(fpath string) bool {
	fi, err := os.Stat(fpath)
	return err == nil && fi.Mode().IsDir()
}
