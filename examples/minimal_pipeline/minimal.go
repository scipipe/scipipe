package main

import . "github.com/scipipe/scipipe"

func main() {
	wfl := NewWorkflow("minimalwf")

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	foo := NewFromShell("fooer", "echo foo > {o:foo}")
	foo.SetPathStatic("foo", "foo.txt")
	wfl.Add(foo)

	f2b := NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetPathExtend("foo", "bar", ".bar.txt")
	wfl.Add(f2b)

	snk := NewSink("sink")
	wfl.SetDriver(snk)

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------

	f2b.In("foo").Connect(foo.Out("foo"))
	snk.Connect(f2b.Out("bar"))

	// --------------------------------
	// Run the workflow!
	// --------------------------------

	wfl.Run()
}
