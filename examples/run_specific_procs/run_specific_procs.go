package main

import (
	"github.com/scipipe/scipipe"
)

func main() {
	wf := scipipe.NewWorkflow("configurable_final_proc", 4)

	first := wf.NewProc("hej_writer", "echo hej > {o:hej}")
	first.SetPathStatic("hej", "hej.txt")

	copyer := wf.NewProc("copyer", "cat {i:in} > {o:out}")
	copyer.SetPathReplace("in", "out", ".txt", ".copy.txt")
	copyer.In("in").Connect(first.Out("hej"))

	wf.RunToProcNames("hej_writer")
}
