package main

import sp "github.com/scipipe/scipipe"

func main() {
	sp.InitLogAudit()
	// Initialize processes
	fooWriter := sp.NewFromShell("fooer", "echo foo > {o:foo}")
	fooWriter.SetPathStatic("foo", "foo.txt")

	fooToBarReplacer := sp.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	fooToBarReplacer.SetPathExtend("foo", "bar", ".bar.txt")

	aSink := sp.NewSink()

	// Connect workflow dependency network
	// ... from out-ports to in-ports ...
	sp.ConnectToFrom(fooToBarReplacer.InPorts["foo"], fooWriter.OutPorts["foo"])
	aSink.Connect(fooToBarReplacer.OutPorts["bar"])

	// Set up a pipeline runner and run the pipeline ...
	pipeRunner := sp.NewPipelineRunner()
	pipeRunner.AddProcesses(fooWriter, fooToBarReplacer, aSink)
	pipeRunner.Run()
}
