package main

import (
	sp "github.com/samuell/scipipe"
)

func main() {
	sp.InitLogDebug()

	t := sp.Sh("ls -l > {o:out}")
	t.OutPathFuncs["out"] = func() string {
		return "hej.txt"
	}
	t.Prepend = "echo"
	t.Run()
}
