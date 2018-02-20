package main

import (
	sci "github.com/scipipe/scipipe"
)

func main() {
	wf := sci.NewWorkflow("filegenwf", 4)

	fq := sci.NewFileIPGenerator(wf, "hej1.txt", "hej2.txt", "hej3.txt")

	fw := sci.NewProc(wf, "filewriter", "echo {i:in} > {o:out}")
	fw.SetPathCustom("out", func(t *sci.Task) string { return t.InPath("in") })
	fw.In("in").Connect(fq.Out())

	wf.Run()
}
