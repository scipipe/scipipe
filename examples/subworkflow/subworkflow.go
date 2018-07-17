// An example that shows how to create a sub-network / sub-workflow that can be
// used as a component
package main

import sp "github.com/scipipe/scipipe"

func main() {
	// Main workflow
	wfl := sp.NewWorkflow("foobar_wf", 4)

	// Sub-workflow
	NewFooBarSubWorkflow(wfl, "foobar_subwf")

	// Run
	wfl.Run()
}

// ------------------------------------------------
// FooBarSubWorkflow
// ------------------------------------------------

type FooBarSubWorkflow struct {
	name  string
	Procs map[string]*sp.Process
	Out   *sp.OutPort
}

func NewFooBarSubWorkflow(wf *sp.Workflow, name string) *FooBarSubWorkflow {
	fbn := &FooBarSubWorkflow{
		name:  name,
		Procs: make(map[string]*sp.Process),
	}

	fbn.Procs["foo"] = sp.NewProc(wf, "foo", "echo foo > {o:foo}")
	fbn.Procs["foo"].SetOut("foo", "foo.txt")

	fbn.Procs["f2b"] = sp.NewProc(wf, "f2b", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	fbn.Procs["f2b"].SetOut("bar", "{i:foo|%.txt}.bar.txt")

	// Connect together inner processes
	fbn.Procs["foo"].Out("foo").To(fbn.Procs["f2b"].In("foo"))

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

func (wf *FooBarSubWorkflow) Ready() bool {
	return wf.Out.Ready()
}
