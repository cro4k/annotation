// Copyright none
package main

import (
	"fmt"
	"github.com/cro4k/annotation/command"
	"github.com/cro4k/common/args"
	"os"
)

func main() {
	kv, cmd := args.Parse()
	msg, err := command.Handle(kv, cmd)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
	} else if msg != "" {
		_, _ = fmt.Fprintln(os.Stdout, msg)
	}
}
