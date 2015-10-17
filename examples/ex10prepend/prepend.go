package main

import (
	sp "github.com/samuell/scipipe"
)

func main() {
	sp.InitLogWarn()

	p := sp.Sh("ls -l > {o:out}")
	p.PathGen["out"] = func(p *sp.ShellTask) string {
		return "hej.txt"
	}
	p.Prepend = "echo"
	p.Run()
}
