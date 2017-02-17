package main

import sp "github.com/scipipe/scipipe"

func main() {
	// --------------------------------
	// Set up a pipeline runner
	// --------------------------------

	run := sp.NewPipelineRunner()

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	foo := sp.NewFromShell("fooer",
		"echo foo > {o:foo}")
	foo.SetPathStatic("foo", "/scidata/foo.txt")
	foo.ExecMode = sp.ExecModeK8s
	run.AddProcess(foo)

	f2b := sp.NewFromShell("foo2bar",
		"sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetPathExtend("foo", "bar", ".bar.txt")
	f2b.ExecMode = sp.ExecModeK8s
	run.AddProcess(f2b)

	snk := sp.NewSink()
	run.AddProcess(snk)

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------

	f2b.In["foo"].Connect(foo.Out["foo"])
	snk.Connect(f2b.Out["bar"])

	// --------------------------------
	// Run the pipeline!
	// --------------------------------

	run.Run()
}
