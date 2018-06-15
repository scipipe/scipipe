package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("test_wf", 4)

	p := sp.NewProc(wf, "ls", "ls -l > {o:out}")
	p.SetPathStatic("out", "hej.txt")
	p.Prepend = "echo"

	wf.Run()
}
