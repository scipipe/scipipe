package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("test_wf", 4)

	p := sp.NewProc(wf, "ls", "ls -l > {o:out}")
	p.SetPathCustom("out", func(p *sp.SciTask) string {
		return "hej.txt"
	})
	p.Prepend = "echo"
	snk := sp.NewSink("sink")
	snk.Connect(p.Out("out"))
	go p.Run()
	snk.Run()
}
