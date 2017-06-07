package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
)

func main() {
	fmt.Println("Starting program!")

	// ls processes
	ls := sci.NewFromShell("ls", "ls -l / > {os:lsl}")
	ls.SetPathCustom("lsl", func(tsk *sci.SciTask) string {
		return "lsl.txt"
	})

	// grep process
	grp := sci.NewFromShell("grp", "grep etc {i:in} > {o:grep}")
	grp.SetPathCustom("grep", func(tsk *sci.SciTask) string {
		return tsk.GetInPath("in") + ".grepped.txt"
	})

	// cat process
	ct := sci.NewFromShell("cat", "cat {i:in} > {o:out}")
	ct.SetPathCustom("out", func(tsk *sci.SciTask) string {
		return tsk.GetInPath("in") + ".out.txt"
	})

	// sink
	snk := sci.NewSink()

	// connect network
	grp.In("in").Connect(ls.Out("lsl"))
	ct.In("in").Connect(grp.Out("grep"))
	snk.Connect(ct.Out("out"))

	// run pipeline
	pl := sci.NewPipelineRunner()
	pl.AddProcesses(ls, grp, ct, snk)
	pl.Run()

	fmt.Println("Finished program!")
}
