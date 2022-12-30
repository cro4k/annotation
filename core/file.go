package core

import (
	"fmt"
	"strings"
)

const (
	TypeFunc   = 1
	TypeStruct = 2
)

const (
	elementTpl = `{
		Type: %d,
		Name: "%s",
		Ptr: %s,
		Path: "%s",
		Comments: []string{%s},
		Annotations: []*core.AnnotationItem{
%s
		},
	}`

	itemTpl = `        {
				Raw: "%s",
				Annotation: "%s",
				Relation: []interface{}{%s},
			}`
)

type element struct {
	Type        int
	Name        string
	Path        string
	Comments    []string
	Exported    bool
	Annotations []*annotationItem
}

type annotationItem struct {
	Raw        string
	Annotation string
	Relation   []*relation
}

type relation struct {
	Path string
	Type int
}

func (e *element) Format() (string, []string) {
	var imports []string
	var anns []string
	for _, v := range e.Annotations {
		annName, annImports := v.Format()
		imports = append(imports, annImports...)
		anns = append(anns, annName)
	}
	var comments string
	if len(e.Comments) > 0 {
		comments = fmt.Sprintf("\"%s\"", strings.Join(e.Comments, "\", \""))
	}

	var impt, name = splitImport(e.Path)
	imports = append(imports, impt)

	var ptr string
	switch e.Type {
	case TypeFunc:
		ptr = name
	case TypeStruct:
		ptr = fmt.Sprintf("new(%s)", name)
	}

	text := fmt.Sprintf(elementTpl,
		e.Type,
		e.Name,
		ptr,
		e.Path,
		comments,
		strings.Join(anns, ",\n")+",",
	)
	return text, imports
}

func (i *annotationItem) Format() (string, []string) {
	var rel string
	var imports []string

	for _, re := range i.Relation {
		imp, name := splitImport(re.Path)
		imports = append(imports, imp)
		switch re.Type {
		case TypeFunc:
			rel += fmt.Sprintf("%s, ", name)
		case TypeStruct:
			rel += fmt.Sprintf("new(%s), ", name)
		}
	}
	text := fmt.Sprintf("    "+itemTpl,
		i.Raw,
		i.Annotation,
		rel,
	)
	return text, imports
}

type GoImport struct {
	Path  string
	Alias string
}

type GoFile struct {
	Path            string      // 文件路径
	Filename        string      // 文件名
	Package         string      // 包名
	ImportPath      string      // 导包路径
	Imports         []*GoImport // 导包列表
	PackageComments []string    // 包注释
	Elements        []*element  //
}

type context struct {
	imports    map[string]string
	pkg        string
	importPath string
}

func (f *GoFile) context() *context {
	ctx := &context{}
	ctx.pkg = f.Package
	ctx.importPath = f.ImportPath
	ctx.imports = make(map[string]string)
	for _, v := range f.Imports {
		if v.Alias == "." || v.Alias == "_" {
			continue
		}
		name := v.Alias
		if name == "" {
			n := strings.LastIndex(v.Path, "/")
			if n >= 0 {
				name = v.Path[n+1:]
			} else {
				name = v.Path
			}
		}
		ctx.imports[name] = v.Path
	}
	return ctx
}

func boolText(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func splitImport(path string) (string, string) {
	var name string
	if n := strings.LastIndex(path, "/"); n > 0 {
		name = path[n+1:]
	} else {
		name = path
	}
	var imp string
	if n := strings.LastIndex(path, "."); n > 0 {
		imp = path[:n]
	} else {
		imp = path
	}
	return imp, name
}
