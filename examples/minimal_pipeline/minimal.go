package main

import sp "github.com/scipipe/scipipe"

func main() {
	rnr := sp.NewPipelineRunner()

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	foo := sp.NewFromShell("fooer", "echo foo > {o:foo}")
	foo.SetPathStatic("foo", "foo.txt")
	rnr.AddProcess(foo)

	f2b := sp.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetPathExtend("foo", "bar", ".bar.txt")
	rnr.AddProcess(f2b)

	snk := sp.NewSink()
	rnr.AddProcess(snk)

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------

	f2b.In("foo").Connect(foo.Out("foo"))
	snk.Connect(f2b.Out("bar"))

	// --------------------------------
	// rnr the pipeline!
	// --------------------------------

	rnr.Run()
}
