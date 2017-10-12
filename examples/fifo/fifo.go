package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
)

func main() {
	fmt.Println("Starting program!")
	wf := sp.NewWorkflow("fifowf", 4)

	// lsl processes
	lsl := sp.NewProc(wf, "lsl", "ls -l / > {os:lsl}")
	lsl.SetPathCustom("lsl", func(tsk *sp.SciTask) string {
		return "lsl.txt"
	})

	// grep process
	grp := sp.NewProc(wf, "grp", "grep etc {i:in} > {o:grep}")
	grp.SetPathCustom("grep", func(tsk *sp.SciTask) string {
		return tsk.InPath("in") + ".grepped.txt"
	})

	// cat process
	cat := sp.NewProc(wf, "cat", "cat {i:in} > {o:out}")
	cat.SetPathCustom("out", func(tsk *sp.SciTask) string {
		return tsk.InPath("in") + ".out.txt"
	})

	// connect network
	grp.In("in").Connect(lsl.Out("lsl"))
	cat.In("in").Connect(grp.Out("grep"))
	wf.ConnectLast(cat.Out("out"))

	// run pipeline
	wf.Run()

	fmt.Println("Finished program!")
}
