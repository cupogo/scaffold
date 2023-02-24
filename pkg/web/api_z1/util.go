package apiz1

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func listModels(dir string) (module string, pkgs []string, err error) {
	f, err := os.Open(path.Join(dir, "go.mod"))
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	if a, b, ok := strings.Cut(scanner.Text(), " "); ok && a == "module" {
		module = b
	}
	if err = scanner.Err(); err != nil {
		return
	}
	pkgs = listSubDirs(dir, "models", "pkg/models")

	return
}

func listSubDirs(dir string, paths ...string) (pkgs []string) {
	for _, path := range paths {
		entries, err := os.ReadDir(filepath.Join(dir, path))
		if err != nil {
			// logger().Debugw("not found", "padh", path)
			continue
		}
		for _, ent := range entries {
			if ent.IsDir() {
				pkgs = append(pkgs, path+"/"+ent.Name())
			}
		}
	}
	return
}
