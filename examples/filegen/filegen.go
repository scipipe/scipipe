package main

import (
	sci "github.com/scipipe/scipipe"
)

func main() {
	fq := sci.NewIPGen("hej1.txt", "hej2.txt", "hej3.txt")
	fw := sci.NewFromShell("filewriter", "echo {i:in} > {o:out}")
	fw.SetPathCustom("out", func(t *sci.SciTask) string { return t.GetInPath("in") })
	sn := sci.NewSink()

	fw.In("in").Connect(fq.Out)
	sn.Connect(fw.Out("out"))

	pl := sci.NewPipelineRunner()
	pl.AddProcesses(fq, fw, sn)

	pl.Run()
}
