package main

import "github.com/scipipe/scipipe"

func main() {
	// Initialize processes
	fooWriter := scipipe.Shell("fooer", "echo foo > {o:foo}")
	fooWriter.SetPathFormatStatic("foo", "foo.txt")

	fooToBarReplacer := scipipe.Shell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	fooToBarReplacer.SetPathFormatExtend("foo", "bar", ".bar.txt")

	aSink := scipipe.NewSink()

	// Connect workflow dependency network
	// ... from out-ports to in-ports ...
	fooToBarReplacer.InPorts["foo"] = fooWriter.OutPorts["foo"]
	aSink.In = fooToBarReplacer.OutPorts["bar"]

	// Set up a pipeline runner and run the pipeline ...
	pipeRunner := scipipe.NewPipelineRunner()
	pipeRunner.AddProcs(fooWriter, fooToBarReplacer, aSink)
	pipeRunner.Run()
}
