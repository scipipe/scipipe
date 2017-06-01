package main

import (
	sci "github.com/scipipe/scipipe"
)

func main() {
	sci.InitLogDebug()

	foo := NewFooer()
	f2b := NewFoo2Barer()
	snk := sci.NewSink()

	f2b.InFoo.Connect(foo.OutFoo)
	snk.Connect(f2b.OutBar)

	pl := sci.NewPipelineRunner()
	pl.AddProcesses(foo, f2b, snk)
	pl.Run()
}

// ------------------------------------------------------------------------
// Components
// ------------------------------------------------------------------------

// Fooer

type Fooer struct {
	InnerProc *sci.SciProcess
	OutFoo    *sci.FilePort
}

func NewFooer() *Fooer {
	innerFoo := sci.NewFromShell("fooer", "echo foo > {o:foo}")
	innerFoo.SetPathStatic("foo", "foo.txt")
	return &Fooer{
		InnerProc: innerFoo,
		OutFoo:    sci.NewFilePort(),
	}
}

func (p *Fooer) Run() {
	p.InnerProc.SetOutPort("foo", p.OutFoo)
	p.InnerProc.Run()
}

func (p *Fooer) IsConnected() bool {
	return p.OutFoo.IsConnected()
}

// Foo2Barer

type Foo2Barer struct {
	InnerProc *sci.SciProcess
	InFoo     *sci.FilePort
	OutBar    *sci.FilePort
}

func NewFoo2Barer() *Foo2Barer {
	innerFoo2Bar := sci.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	innerFoo2Bar.SetPathExtend("foo", "bar", ".bar.txt")
	return &Foo2Barer{
		InnerProc: innerFoo2Bar,
		InFoo:     sci.NewFilePort(),
		OutBar:    sci.NewFilePort(),
	}
}

func (p *Foo2Barer) Run() {
	p.InnerProc.SetInPort("foo", p.InFoo)
	p.InnerProc.SetOutPort("bar", p.OutBar)
	p.InnerProc.Run()
}

func (p *Foo2Barer) IsConnected() bool {
	return p.InFoo.IsConnected() &&
		p.OutBar.IsConnected()
}
