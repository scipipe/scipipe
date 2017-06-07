// An example that shows how to create a sub-network / sub-workflow that can be
// used as a component
package main

import sp "github.com/scipipe/scipipe"

func main() {
	// Main workflow
	wfl := sp.NewWorkflow("foobar_wf")

	// Sub-workflow
	fbn := NewFooBarSubWorkflow("foobar_subwf")
	wfl.Add(fbn)

	// Sink
	snk := sp.NewSink("sink")
	wfl.SetDriver(snk)

	// Connect
	snk.Connect(fbn.Out)

	wfl.Run()
}

// ------------------------------------------------
// FooBarSubWorkflow
// ------------------------------------------------

type FooBarSubWorkflow struct {
	sp.Process
	name  string
	Procs map[string]*sp.SciProcess
	Out   *sp.FilePort
}

func NewFooBarSubWorkflow(name string) *FooBarSubWorkflow {
	fbn := &FooBarSubWorkflow{
		name:  name,
		Procs: make(map[string]*sp.SciProcess),
	}

	fbn.Procs["foo"] = sp.NewProc("foo", "echo foo > {o:foo}")
	fbn.Procs["foo"].SetPathStatic("foo", "foo.txt")

	fbn.Procs["f2b"] = sp.NewProc("f2b", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	fbn.Procs["f2b"].SetPathReplace("foo", "bar", ".txt", ".bar.txt")

	// Connect together inner processes
	fbn.Procs["foo"].Out("foo").Connect(fbn.Procs["f2b"].In("foo"))

	// Connect last port of inner process to subnetwork out-port
	fbn.Out = fbn.Procs["f2b"].Out("bar")

	return fbn
}

func (wf *FooBarSubWorkflow) Name() string {
	return wf.name
}

func (wf *FooBarSubWorkflow) Run() {
	for _, proc := range wf.Procs {
		go proc.Run()
	}
}

func (wf *FooBarSubWorkflow) IsConnected() bool {
	return wf.Out.IsConnected()
}
