package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	sp.InitLogWarning()

	p := sp.Shell("ls", "ls -l > {o:out}")
	p.PathFormatters["out"] = func(p *sp.SciTask) string {
		return "hej.txt"
	}
	p.Prepend = "echo"
	p.Run()
}
