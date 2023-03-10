package core

import (
	"fmt"
	"github.com/cro4k/annotation/version"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const tpl = `// Code generated by annotation. DO NOT EDIT.
// version: {{version}}

package {{package}}

import(
{{imports}}
)

const (
	TypeFunc   = 1
	TypeStruct = 2
)

var elements = map[string]*core.Element{
{{elements}}
}

func Elements() map[string]*core.Element {
	return elements
}

`

const (
	dotAnnotationFile    = ".annotation"
	dotAnnotationContent = "This directory is auto generated by annotation, do not edit/create/remove any file in this directory."
)

func isAnnotationDir(path string) bool {
	_, err := os.Stat(fmt.Sprintf("%s/%s", strings.TrimRight(path, "/"), dotAnnotationFile))
	return err == nil
}

func Export(path string) error {
	if path == "" {
		path = "annotation"
	}
	path = strings.TrimRight(path, "/")
	if _, err := os.Stat(path); err == nil {
		if isAnnotationDir(path) {
			os.RemoveAll(path)
		} else {
			return fmt.Errorf("path '%s' has been used by another package", path)
		}
	}
	os.MkdirAll(path, 0777)
	os.WriteFile(fmt.Sprintf("%s/%s", path, dotAnnotationFile), []byte(dotAnnotationContent), 0644)
	var pkg string
	if n := strings.LastIndex(path, "/"); n > 0 {
		pkg = path[n+1:]
	} else {
		pkg = path
	}

	mod, err := decodeMod()
	if err != nil {
		return err
	}

	files, err := decodeGoFiles(mod.Module, ".")
	if err != nil {
		return err
	}

	var body []string
	var imports = map[string]struct{}{"github.com/cro4k/annotation/core": {}}
	for _, v := range files {
		for _, ele := range v.Elements {
			if !ele.Exported || len(ele.Annotations) == 0 {
				continue
			}
			format, imps := ele.Format()
			for _, im := range imps {
				imports[im] = struct{}{}
			}
			body = append(body, fmt.Sprintf("    \"%s\": ", ele.Path)+format)
		}
	}

	content := strings.ReplaceAll(tpl, "{{package}}", pkg)
	content = strings.ReplaceAll(content, "{{version}}", version.Version)
	if len(body) > 0 {
		content = strings.ReplaceAll(content, "{{elements}}", strings.Join(body, ",\n")+",")
	} else {
		content = strings.ReplaceAll(content, "{{elements}}", "")
	}
	var imps string
	for imp := range imports {
		imps += fmt.Sprintf("    \"%s\"\n", imp)
	}
	content = strings.ReplaceAll(content, "{{imports}}", imps)

	filename := fmt.Sprintf("%s/annotation.go", path)
	return os.WriteFile(filename, []byte(content), 0644)
}

// Clean
// remove all files generated by annotation
func Clean() error {
	return filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if path == "." {
			return nil
		}
		if info.IsDir() {
			if _, e := os.Stat(fmt.Sprintf("%s/%s", path, dotAnnotationFile)); e == nil {
				if er := os.RemoveAll(path); er != nil {
					return er
				}
				return filepath.SkipDir
			}
		}
		return err
	})
}
