package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	p := sp.NewProc("ls", "ls -l > {o:out}")
	p.SetPathCustom("out", func(p *sp.SciTask) string {
		return "hej.txt"
	})
	p.Prepend = "echo"
	snk := sp.NewSink("sink")
	snk.Connect(p.Out("out"))
	go p.Run()
	snk.Run()
}
