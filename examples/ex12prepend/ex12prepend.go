package main

import (
	sp "github.com/samuell/scipipe"
)

func main() {
	sp.InitLogInfo()

	t := sp.Sh("ls -l > out.txt")
	t.Prepend = "echo"
	t.Run()
}
