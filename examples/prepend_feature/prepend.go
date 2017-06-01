package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	sp.InitLogAudit()

	p := sp.NewFromShell("ls", "ls -l > {o:out}")
	p.SetPathCustom("out", func(p *sp.SciTask) string {
		return "hej.txt"
	})
	p.Prepend = "echo"
	snk := sp.NewSink()
	snk.Connect(p.Out("out"))
	go p.Run()
	snk.Run()
}
