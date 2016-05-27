package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	sp.InitLogAudit()

	p := sp.Shell("ls", "ls -l > {o:out}")
	p.PathFormatters["out"] = func(p *sp.SciTask) string {
		return "hej.txt"
	}
	p.Prepend = "echo"
	snk := sp.NewSink()
	snk.Connect(p.OutPorts["out"])
	go p.Run()
	snk.Run()
}
