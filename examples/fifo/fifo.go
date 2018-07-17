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
	lsl.SetOut("lsl", "lsl.txt")

	// grep process
	grp := sp.NewProc(wf, "grp", "grep etc {i:in} > {o:grep}")
	grp.SetOut("grep", "{i:in}.grepped.txt")

	// cat process
	cat := sp.NewProc(wf, "cat", "cat {i:in} > {o:out}")
	cat.SetOut("out", "{i:in}.out.txt")

	// connect network
	grp.In("in").From(lsl.Out("lsl"))
	cat.In("in").From(grp.Out("grep"))

	// run pipeline
	wf.Run()

	fmt.Println("Finished program!")
}
