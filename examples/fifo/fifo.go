package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
)

func main() {
	fmt.Println("Starting program!")

	// lsl processes
	lsl := sp.NewProc("lsl", "ls -l / > {os:lsl}")
	lsl.SetPathCustom("lsl", func(tsk *sp.SciTask) string {
		return "lsl.txt"
	})

	// grep process
	grp := sp.NewProc("grp", "grep etc {i:in} > {o:grep}")
	grp.SetPathCustom("grep", func(tsk *sp.SciTask) string {
		return tsk.GetInPath("in") + ".grepped.txt"
	})

	// cat process
	cat := sp.NewProc("cat", "cat {i:in} > {o:out}")
	cat.SetPathCustom("out", func(tsk *sp.SciTask) string {
		return tsk.GetInPath("in") + ".out.txt"
	})

	// sink
	snk := sp.NewSink("sink")

	// connect network
	grp.In("in").Connect(lsl.Out("lsl"))
	cat.In("in").Connect(grp.Out("grep"))
	snk.Connect(cat.Out("out"))

	// run pipeline
	wf := sp.NewWorkflow("fifowf")
	wf.AddProcs(lsl, grp, cat)
	wf.SetDriver(snk)
	wf.Run()

	fmt.Println("Finished program!")
}
