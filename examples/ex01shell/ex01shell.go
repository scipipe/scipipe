package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogInfo()

	h := sci.Sh("echo foo > {o:foo}")
	h.OutPathFuncs["foo"] = func() string {
		return "foo.txt"
	}

	f2b := sci.Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.OutPathFuncs["bar"] = func() string {
		return fmt.Sprint(f2b.GetInPath("foo"), ".bar")
	}

	p := sci.Sh("cat {i:inf}")

	f2b.InPorts["foo"] = h.OutPorts["foo"]
	p.InPorts["inf"] = f2b.OutPorts["bar"]

	go h.Run()
	go f2b.Run()
	p.Run()
}
