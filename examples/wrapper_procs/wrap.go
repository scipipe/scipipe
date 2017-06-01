package main

import (
	sci "github.com/scipipe/scipipe"
)

func main() {
	sci.InitLogInfo()
	plr := sci.NewPipelineRunner()

	// Fooer
	foo := NewFooer()
	plr.AddProcess(foo)

	// Foo2barer
	f2b := NewFoo2Barer()
	f2b.InFoo().Connect(foo.OutFoo())
	plr.AddProcess(f2b)

	// Sink
	snk := sci.NewSink()
	snk.Connect(f2b.OutBar())
	plr.AddProcess(snk)

	// Run
	plr.Run()
}

// ------------------------------------------------
// Components
// ------------------------------------------------

// Fooer
// -----

type Fooer struct {
	*sci.SciProcess
}

func NewFooer() *Fooer {
	innerFoo := sci.NewFromShell("fooer", "echo foo > {o:foo}")
	innerFoo.SetPathStatic("foo", "foo.txt")
	return &Fooer{innerFoo}
}

// Define static ports

func (p *Fooer) OutFoo() *sci.FilePort {
	return p.Out("foo")
}

// Foo2Barer
// ---------

type Foo2Barer struct {
	*sci.SciProcess
}

func NewFoo2Barer() *Foo2Barer {
	innerFoo2Bar := sci.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	innerFoo2Bar.SetPathExtend("foo", "bar", ".bar.txt")
	return &Foo2Barer{innerFoo2Bar}
}

// Define static ports

func (p *Foo2Barer) InFoo() *sci.FilePort {
	return p.In("foo")
}

func (p *Foo2Barer) OutBar() *sci.FilePort {
	return p.Out("bar")
}
