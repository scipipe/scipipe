package main

import sp "github.com/scipipe/scipipe"

func main() {
	// Initialize workflow
	wf := sp.NewWorkflow("minimal_wf", 4)

	// Initialize processes
	foo := wf.NewProc("fooer", "echo foo > {o:foo.txt}")
	f2b := wf.NewProc("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar.txt}")

	// Connect
	f2b.In("foo").From(foo.Out("foo"))

	// Run
	wf.Run()
}
