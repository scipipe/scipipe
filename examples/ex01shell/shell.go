package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogInfo()

	h := sci.Sh("echo foo > {o:foo}")
	h.PathFormatters["foo"] = func(t *sci.ShellTask) string {
		return "foo.txt"
	}

	f2b := sci.Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.PathFormatters["bar"] = func(t *sci.ShellTask) string {
		return fmt.Sprint(t.GetInPath("foo"), ".bar")
	}

	sn := sci.NewSink()

	f2b.InPorts["foo"] = h.OutPorts["foo"]
	sn.In = f2b.OutPorts["bar"]

	pl := sci.NewPipeline()
	pl.AddProcs(h, f2b, sn)
	pl.Run()
}
