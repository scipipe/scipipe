package main

import (
	sci "github.com/samuell/scipipe"
)

func main() {
	barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo") + ".bar"
	}
	barReplacer.Init()

	for _, name := range []string{"foo1.txt", "foo2.txt", "foo3.txt"} {
		barReplacer.InPorts["foo"] <- sci.NewFileTarget(name)
	}
	close(barReplacer.InPorts["foo"])

	for {
		<-barReplacer.OutPorts["bar"]
	}
}
