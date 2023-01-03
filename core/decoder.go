package core

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func decodeFile(module, filename string) (*GoFile, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	re := bufio.NewReader(fi)

	var file = new(GoFile)
	file.Filename = strings.ReplaceAll(filename, "\\", "/")

	if n := strings.LastIndex(file.Filename, "/"); n >= 0 {
		file.Path = file.Filename[:n]
	} else {
		file.Path = "."
	}

	var comments []string
	for {
		line, err := readLine(re)
		if err != nil {
			break
		}
		if !strings.HasPrefix(line, "/") {
			line = trimInlineComment(line)
		}
		if strings.Contains(line, "`") {
			line, err = trimMultiLine(line, re)
			if err != nil {
				break
			}
		}

		if strings.HasPrefix(line, "package") {
			file.Package = strings.TrimSpace(strings.TrimPrefix(line, "package"))
			file.PackageComments = comments
			file.ImportPath = fmt.Sprintf("%s/%s", module, file.Path)
			//if file.Package != "main" {
			//	file.ImportPath = fmt.Sprintf("%s/%s", module, file.Path)
			//} else {
			//	file.ImportPath = "main"
			//}
			comments = []string{}
			continue
		}

		if strings.HasPrefix(line, "import") {
			file.Imports = append(file.Imports, readImports(line, re)...)
			comments = []string{}
			continue
		}
		if strings.HasPrefix(line, "func") || strings.HasPrefix(line, "type") {
			e := readFuncOrStruct(line, re, comments)
			e.Path = fmt.Sprintf("%s.%s", file.ImportPath, e.Name)
			file.Elements = append(file.Elements, e)
			comments = []string{}
			continue
		}

		var isComment = false
		if strings.HasPrefix(line, "/*") {
			isComment = true
			comments = append(comments, readComments(line, re)...)
		}
		if strings.HasPrefix(line, "//") {
			isComment = true
			comments = append(comments, strings.TrimSpace(strings.TrimPrefix(line, "//")))
		}

		if !isComment {
			comments = []string{}
		}

	}
	return file, nil
}

func readComments(lastLine string, re *bufio.Reader) []string {
	if strings.HasSuffix(lastLine, "*/") {
		return []string{strings.Trim(lastLine, "/*")}
	}
	var comments []string
	if theFirst := strings.TrimLeft(lastLine, "/*"); theFirst != "" {
		comments = append(comments, theFirst)
	}
	for {
		line, err := readLine(re)
		if err != nil {
			return comments
		}
		if strings.HasSuffix(line, "*/") {
			if theLast := strings.TrimRight(line, "*/"); theLast != "" {
				comments = append(comments, theLast)
			}
			return comments
		} else {
			comments = append(comments, line)
		}
	}
}

func readImports(lastLine string, re *bufio.Reader) []*GoImport {
	if !strings.Contains(lastLine, "(") {
		return []*GoImport{readImport(lastLine)}
	}
	var imps []*GoImport
	for {
		line, err := readLine(re)
		if err != nil {
			return imps
		}
		if line == "" {
			continue
		}
		if line == ")" {
			return imps
		}
		imps = append(imps, readImport(lastLine))
	}
}

func readImport(line string) *GoImport {
	temp := strings.Fields(line)
	var im = new(GoImport)
	if len(temp) == 2 {
		im.Path = strings.Trim(temp[1], "\"")
	} else if len(temp) == 3 {
		im.Path = strings.Trim(temp[2], "\"")
		im.Alias = []string{temp[1]}
	}
	if replace := config.Replace[im.Path]; len(replace) > 0 {
		im.Alias = append(im.Alias, replace...)
	}
	return im
}

func readFuncOrStruct(lastLine string, re *bufio.Reader, comments []string) *element {
	lastLine = trimInlineComment(lastLine)
	funcFields := strings.Fields(lastLine)
	var name = funcFields[1]
	if strings.HasPrefix(lastLine, "func") {
		if name[0] == '(' {
			name = funcFields[3]
		}
		name = strings.Split(name, "(")[0]
	}

	var e = new(element)
	e.Name = name
	e.Comments = comments
	e.Exported = name[0] >= 'A' && name[0] <= 'Z'
	if strings.HasPrefix(lastLine, "func") {
		e.Type = TypeFunc
	} else if strings.HasPrefix(lastLine, "type") {
		e.Type = TypeStruct
	}

	var left, right int
	left = strings.Count(lastLine, "{")
	right = strings.Count(lastLine, "}")

	for left != right {
		line, err := readLine(re)
		if err != nil {
			break
		}
		left += strings.Count(line, "{")
		right += strings.Count(line, "}")
	}
	return e
}

func readLine(re *bufio.Reader) (string, error) {
	data, err := re.ReadBytes('\n')
	if err != nil || len(data) == 0 {
		return "", errors.New("EOF")
	}
	line := strings.TrimSpace(string(data))
	return line, nil
}

func trimInlineComment(line string) string {
	var quotes bool
	var slash bool
	var apostrophe bool
	for i, v := range line {
		if v == '/' {
			if quotes || apostrophe {
				continue
			} else if slash {
				return strings.TrimSpace(line[:i-1])
			} else {
				slash = true
			}
		} else {
			slash = false
		}
		if v == '"' {
			quotes = !quotes
		}
		if v == '`' && !quotes {
			apostrophe = !apostrophe
		}
	}
	return line
}

func trimMultiLine(line string, re *bufio.Reader) (string, error) {
	count := strings.Count(line, "`")
	if count%2 == 0 {
		return line, nil
	}
	for {
		next, err := readLine(re)
		if err != nil {
			return next, err
		}
		count += strings.Count(next, "`")
		if count%2 == 0 {
			return readLine(re)
		}
	}
}

var reg, _ = regexp.Compile("\\[[\\w.]+\\]|<[\\w.]+>")

func decodeAnnotations(ctx *context, e *element) {
	for _, line := range e.Comments {
		if !strings.HasPrefix(line, "@") {
			continue
		}
		i := &annotationItem{Raw: line[1:]}
		i.Annotation = i.Raw
		for _, v := range reg.FindAllString(line, -1) {
			rel := strings.Trim(strings.Trim(v, "[]"), "<>")
			n := strings.LastIndex(rel, ".")
			var relationPath string
			if n > 0 {
				prefix := rel[:n]
				name := rel[n+1:]
				if p := ctx.imports[prefix]; p != "" {
					relationPath = fmt.Sprintf("%s.%s", p, name)
				} else {
					relationPath = rel
				}
			} else {
				relationPath = fmt.Sprintf("%s.%s", ctx.importPath, rel)
			}
			re := &relation{Path: relationPath}
			if strings.HasPrefix(v, "[") {
				re.Type = TypeStruct
			} else if strings.HasPrefix(v, "<") {
				re.Type = TypeFunc
			}
			i.Relation = append(i.Relation, re)
			i.Annotation = strings.ReplaceAll(i.Annotation, v, v[0:1]+relationPath+v[len(v)-1:])
		}
		e.Annotations = append(e.Annotations, i)
	}
}

func decodeGoFiles(module string, root string) ([]*GoFile, error) {
	var files []*GoFile
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if path == root {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		file, err := decodeFile(module, path)
		if err != nil {
			return err
		}

		for _, e := range file.Elements {
			decodeAnnotations(file.context(), e)
		}
		files = append(files, file)
		return err
	})
	return files, err
}

type GoMod struct {
	Module    string
	GoVersion string
}

func decodeMod() (*GoMod, error) {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return nil, err
	}
	var mod = new(GoMod)
	for _, v := range strings.Split(string(b), "\n") {
		line := strings.TrimSpace(v)
		if strings.HasPrefix(line, "module") {
			mod.Module = strings.TrimSpace(strings.TrimPrefix(line, "module"))
		} else if strings.HasPrefix(line, "go") {
			mod.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go"))
		}
	}
	return mod, nil
}
