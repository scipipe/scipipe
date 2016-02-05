package main

import (
	sp "github.com/samuell/scipipe"
)

func main() {
	sp.InitLogWarning()

	t := sp.Sh("ls -l > {o:out}")
	t.PathFormatters["out"] = func(t *sp.ShellTask) string {
		return "hej.txt"
	}
	t.Prepend = "echo"
	t.Run()
}
