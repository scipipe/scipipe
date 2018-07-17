package main

import (
	// Import the SciPipe package, aliased to 'sp'
	sp "github.com/scipipe/scipipe"
)

func main() {
	// Init workflow with a name, and max concurrent tasks
	wf := sp.NewWorkflow("hello_world", 4)

	// Initialize processes and set output file paths
	hello := wf.NewProc("hello", "echo 'Hello ' > {o:out}")
	world := wf.NewProc("world", "echo $(cat {i:in}) World >> {o:out}")

	// Connect network
	world.In("in").From(hello.Out("out"))

	// Run workflow
	wf.Run()
}
