package main

import (
	sci "github.com/samuell/scipipe"
)

func main() {
	// Init fooWriter task
	fooWriter := sci.Sh("echo foo > {o:foo1}")
	// Init function for generating output file pattern
	fooWriter.OutPathFuncs["foo1"] = func() string {
		return "foo.txt"
	}

	// Init barReplacer task
	barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo2} > {o:bar}")
	// Init function for generating output file pattern
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo2") + ".bar"
	}

	// Connect network
	barReplacer.InPorts["foo2"] = fooWriter.OutPorts["foo1"]

	// Set up tasks for execution
	fooWriter.Init()
	barReplacer.Init()

	// Start execution by reading on last port
	<-barReplacer.OutPorts["bar"]
}
