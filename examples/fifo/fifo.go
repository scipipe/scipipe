package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
)

func main() {
	sci.InitLogAudit()

	fmt.Println("Starting program!")

	// ls processes
	ls := sci.NewFromShell("ls", "ls -l / > {os:lsl}")
	ls.PathFormatters["lsl"] = func(tsk *sci.SciTask) string {
		return "lsl.txt"
	}

	// grep process
	grp := sci.NewFromShell("grp", "grep etc {i:in} > {o:grep}")
	grp.PathFormatters["grep"] = func(tsk *sci.SciTask) string {
		return tsk.GetInPath("in") + ".grepped.txt"
	}

	// cat process
	ct := sci.NewFromShell("cat", "cat {i:in} > {o:out}")
	ct.PathFormatters["out"] = func(tsk *sci.SciTask) string {
		return tsk.GetInPath("in") + ".out.txt"
	}

	// sink
	snk := sci.NewSink()

	// connect network
	grp.InPorts["in"].Connect(ls.OutPorts["lsl"])
	ct.InPorts["in"].Connect(grp.OutPorts["grep"])
	snk.Connect(ct.OutPorts["out"])

	// run pipeline
	pl := sci.NewPipelineRunner()
	pl.AddProcesses(ls, grp, ct, snk)
	pl.Run()

	fmt.Println("Finished program!")
}
