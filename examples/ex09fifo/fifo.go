package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
)

func main() {
	sci.InitLogAudit()

	fmt.Println("Starting program!")

	// ls processes
	ls := sci.Shell("ls", "ls -l / > {os:lsl}")
	ls.PathFormatters["lsl"] = func(tsk *sci.SciTask) string {
		return "lsl.txt"
	}

	// grep process
	grp := sci.Shell("grp", "grep etc {i:in} > {o:grep}")
	grp.PathFormatters["grep"] = func(tsk *sci.SciTask) string {
		return tsk.GetInPath("in") + ".grepped.txt"
	}

	// cat process
	ct := sci.Shell("cat", "cat {i:in} > {o:out}")
	ct.PathFormatters["out"] = func(tsk *sci.SciTask) string {
		return tsk.GetInPath("in") + ".out.txt"
	}

	// sink
	snk := sci.NewSink()

	// connect network
	grp.InPorts["in"] = ls.OutPorts["lsl"]
	ct.InPorts["in"] = grp.OutPorts["grep"]
	snk.In = ct.OutPorts["out"]

	// run pipeline
	pl := sci.NewPipelineRunner()
	pl.AddProcesses(ls, grp, ct, snk)
	pl.Run()

	fmt.Println("Finished program!")
}
