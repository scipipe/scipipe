package main

import "github.com/scipipe/scipipe"

func main() {
	pl := scipipe.NewPipeline()

	// ---------------------------------------------------------------
	// Initialize processes from shell command patterns
	// ---------------------------------------------------------------

	pl.NewProc("foo_writer", "echo foo > {o:foo}")
	pl.GetProc("foo_writer").SetPathStatic("foo", "foo.txt")

	pl.NewProc("foo_to_bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	pl.GetProc("foo_to_bar").SetPathExtend("foo", "bar", ".bar.txt")

	// ---------------------------------------------------------------
	// Connect data flow network
	// ---------------------------------------------------------------

	pl.Connect("foo_to_bar.foo <- foo_writer.foo")

	// ---------------------------------------------------------------
	// Run pipeline
	// ---------------------------------------------------------------

	pl.Run()
}
