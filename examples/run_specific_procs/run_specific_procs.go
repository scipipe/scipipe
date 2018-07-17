package main

import (
	"github.com/scipipe/scipipe"
)

func main() {
	wf := scipipe.NewWorkflow("configurable_final_proc", 4)

	first := wf.NewProc("hej_writer", "echo hej > {o:hej}")
	first.SetOut("hej", "hej.txt")

	copyer := wf.NewProc("copyer", "cat {i:in} > {o:out}")
	copyer.SetOut("out", "{i:in|%.txt}.copy.txt")

	copyer.In("in").From(first.Out("hej"))

	wf.RunTo("hej_writer")
}
