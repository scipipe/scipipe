package main

import sp "github.com/scipipe/scipipe"

func main() {
	// Initialize a new workflow
	wfl := sp.NewWorkflow("minimal_wf", 4)

	// Initialize processes and add to runner
	foo := wfl.NewProc("fooer", "echo foo > {o:foo}")
	f2b := wfl.NewProc("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")

	// Connect workflow dependency network
	f2b.In("foo").From(foo.Out("foo"))

	// Run the workflow!
	wfl.Run()
}
