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
	f2b.InFoo().Connect(foo.OutFoo())
	wfl.ConnectLast(f2b.OutBar())

	// Run
	wfl.Run()
}

// ------------------------------------------------
// Components
// ------------------------------------------------

// Fooer
// -----

type Fooer struct {
	*sci.SciProcess
	name string
}

func NewFooer(wf *sci.Workflow, name string) *Fooer {
	innerFoo := sci.NewProc(wf, "fooer", "echo foo > {o:foo}")
	innerFoo.SetPathStatic("foo", "foo.txt")
	return &Fooer{innerFoo, name}
}

// Define static ports

func (p *Fooer) OutFoo() *sci.FilePort {
	return p.Out("foo")
}

// Foo2Barer
// ---------

type Foo2Barer struct {
	*sci.SciProcess
	name string
}

func NewFoo2Barer(wf *sci.Workflow, name string) *Foo2Barer {
	innerFoo2Bar := sci.NewProc(wf, "foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	innerFoo2Bar.SetPathExtend("foo", "bar", ".bar.txt")
	return &Foo2Barer{innerFoo2Bar, name}
}

// Define static ports

func (p *Foo2Barer) InFoo() *sci.FilePort {
	return p.In("foo")
}

func (p *Foo2Barer) OutBar() *sci.FilePort {
	return p.Out("bar")
}
