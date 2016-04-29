package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	sp.InitLogDebug()
	sp.ExecCmd("echo 'Hello World!' > out.txt")
}
