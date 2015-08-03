package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogDebug()

	fmt.Println("Starting program!")
	ls := sci.Shell("ls -l / > {os:lsl}")
	ls.OutPathFuncs["lsl"] = func(tsk *sci.ShellTask) string {
		return "lsl.txt"
	}

	grp := sci.Shell("grep etc {i:in} > {o:grep}")
	grp.OutPathFuncs["grep"] = func(tsk *sci.ShellTask) string {
		return tsk.GetInPath("in") + ".grepped.txt"
	}

	ct := sci.Shell("cat {i:in} > {o:out}")
	ct.OutPathFuncs["out"] = func(tsk *sci.ShellTask) string {
		return tsk.GetInPath("in") + ".out.txt"
	}

	snk := sci.NewSink()

	grp.InPorts["in"] = ls.OutPorts["lsl"]
	ct.InPorts["in"] = grp.OutPorts["grep"]
	snk.In = ct.OutPorts["out"]

	pl := sci.NewPipeline()
	pl.AddProcs(ls, grp, ct, snk)
	pl.Run()

	fmt.Println("Finished program!")
}
