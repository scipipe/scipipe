package main

import (
	sci "github.com/samuell/scipipe"
)

// ****** Main ******

func main() {
	fooWriter := sci.ShOut("echo foo > {o:foo1}")
	fooWriter.OutPathFuncs["foo1"] = func() string {
		return "foo.txt"
	}

	barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo2} > {o:bar}")
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo2") + ".bar"
	}
	barReplacer.InPorts["foo2"] = fooWriter.OutPorts["foo1"]

	fooWriter.Init()
	barReplacer.Init()

	<-barReplacer.OutPorts["bar"]
}
