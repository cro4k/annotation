package core

type Element struct {
	Type        int
	Name        string
	Ptr         interface{}
	Path        string
	Comments    []string
	Annotations []*AnnotationItem
}

type AnnotationItem struct {
	Raw        string
	Annotation string
	Relation   []interface{}
}
