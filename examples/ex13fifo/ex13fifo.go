package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogWarn()

	fmt.Println("Starting program!")

	// ls processes
	ls := sci.Shell("ls -l / > {os:lsl}")
	ls.PathGen["lsl"] = func(tsk *sci.ShellTask) string {
		return "lsl.txt"
	}

	// grep process
	grp := sci.Shell("grep etc {i:in} > {o:grep}")
	grp.PathGen["grep"] = func(tsk *sci.ShellTask) string {
		return tsk.GetInPath("in") + ".grepped.txt"
	}

	// cat process
	ct := sci.Shell("cat {i:in} > {o:out}")
	ct.PathGen["out"] = func(tsk *sci.ShellTask) string {
		return tsk.GetInPath("in") + ".out.txt"
	}

	// sink
	snk := sci.NewSink()

	// connect network
	grp.InPorts["in"] = ls.OutPorts["lsl"]
	ct.InPorts["in"] = grp.OutPorts["grep"]
	snk.In = ct.OutPorts["out"]

	// run pipeline
	pl := sci.NewPipeline()
	pl.AddProcs(ls, grp, ct, snk)
	pl.Run()

	fmt.Println("Finished program!")
}
