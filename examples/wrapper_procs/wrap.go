package main

import (
	sci "github.com/scipipe/scipipe"
)

func main() {
	// Initiate
	wfl := sci.NewWorkflow("wrapperwf", 4)
	foo := NewFooer(wfl, "fooer")
	f2b := NewFoo2Barer(wfl, "foo2barer")

	// Connect
	f2b.InFoo().From(foo.OutFoo())

	// Run
	wfl.Run()
}

// ------------------------------------------------
// Components
// ------------------------------------------------

// Fooer
// -----

type Fooer struct {
	*sci.Process
	name string
}

func NewFooer(wf *sci.Workflow, name string) *Fooer {
	innerFoo := sci.NewProc(wf, "fooer", "echo foo > {o:foo}")
	innerFoo.SetOut("foo", "foo.txt")
	return &Fooer{innerFoo, name}
}

// Define static ports

func (p *Fooer) OutFoo() *sci.OutPort {
	return p.Out("foo")
}

// Foo2Barer
// ---------

type Foo2Barer struct {
	*sci.Process
	name string
}

func NewFoo2Barer(wf *sci.Workflow, name string) *Foo2Barer {
	innerFoo2Bar := sci.NewProc(wf, "foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	innerFoo2Bar.SetOut("bar", "{i:foo}.bar.txt")
	return &Foo2Barer{innerFoo2Bar, name}
}

// Define static ports

func (p *Foo2Barer) InFoo() *sci.InPort {
	return p.In("foo")
}

func (p *Foo2Barer) OutBar() *sci.OutPort {
	return p.Out("bar")
}
