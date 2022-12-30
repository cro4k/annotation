package example

import "github.com/cro4k/annotation/core"

type Example struct {
	Name    string
	Element *core.Element
}

// Hello
// @req [Example] <fmt.Println>
// @rsp [Example] [core.Element]
// @comment Example
func Hello(e Example) Example {
	return e
}

func SayHello() {

}
