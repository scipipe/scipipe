package main

import . "github.com/scipipe/scipipe"

func main() {
	wfl := NewWorkflow("minimal_wf", 4)

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	foo := wfl.NewProc("fooer", "echo foo > {o:foo}")
	foo.SetPathStatic("foo", "foo.txt")

	f2b := wfl.NewProc("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetPathExtend("foo", "bar", ".bar.txt")

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------

	f2b.In("foo").Connect(foo.Out("foo"))

	// --------------------------------
	// Run the workflow!
	// --------------------------------

	wfl.Run()
}
