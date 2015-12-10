package main

import (
	sp "github.com/samuell/scipipe"
)

func main() {
	sp.InitLogDebug()
	sp.ExecCmd("echo 'Hello World!' > out.txt")
}
