package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	// Init barReplacer task
	barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo2} > {o:bar}")
	// Init function for generating output file pattern
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo2") + ".bar"
	}

	// Set up tasks for execution
	barReplacer.Init()

	// Connect network
	for _, name := range []string{"foo1", "foo2", "foo3"} {
		barReplacer.InPorts["foo2"] <- sci.NewFileTarget(name + ".txt")
	}
	close(barReplacer.InPorts["foo2"])
	for f := range barReplacer.OutPorts["bar"] {
		fmt.Println("Processed file", f.GetPath(), "...")
	}
}
