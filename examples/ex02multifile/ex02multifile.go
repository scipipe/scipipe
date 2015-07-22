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
	go barReplacer.Run()

	// Manually send file targets on the inport of barReplacer
	for _, name := range []string{"foo1", "foo2", "foo3"} {
		barReplacer.InPorts["foo2"] <- sci.NewFileTarget(name + ".txt")
	}
	// We have to manually close the inport as well here, to
	// signal that we are done sending targets (the tasks outport will
	// then automatically be closed as well)
	close(barReplacer.InPorts["foo2"])

	for f := range barReplacer.OutPorts["bar"] {
		fmt.Println("Finished processing file", f.GetPath(), "...")
	}
}
