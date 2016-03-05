package main

import (
	"bytes"

	sci "github.com/samuell/scipipe"
)

func main() {
	foo := NewFooer()
	f2b := NewFoo2Barer()
	snk := sci.NewSink()

	f2b.InFoo = foo.OutFoo
	snk.In = f2b.OutBar

	pl := sci.NewPipeline()
	pl.AddProcs(foo, f2b, snk)
	pl.Run()
}

// ------------------------------------------------------------------------
// Components
// ------------------------------------------------------------------------

// Fooer

type Fooer struct {
	InnerProc *sci.ShellProcess
	OutFoo    chan *sci.FileTarget
}

func NewFooer() *Fooer {
	innerFoo := sci.Shell("{o:foo}")
	innerFoo.SetPathFormatterString("foo", "foo.txt")
	fooer := &Fooer{
		InnerProc: innerFoo,
		OutFoo:    innerFoo.OutPorts["foo"],
	}
	fooer.InnerProc.CustomExecute = func(task *sci.ShellTask) {
		task.OutTargets["foo"].WriteTempFile([]byte("foo\n"))
	}
	return fooer
}

func (p *Fooer) Run() {
	p.InnerProc.OutPorts["foo"] = p.OutFoo
	p.InnerProc.Run()
}

// Foo2Barer

type Foo2Barer struct {
	InnerProc *sci.ShellProcess
	InFoo     chan *sci.FileTarget
	OutBar    chan *sci.FileTarget
}

func NewFoo2Barer() *Foo2Barer {
	innerFoo2Bar := sci.Shell("{i:foo}{o:bar}")
	innerFoo2Bar.SetPathFormatterExtend("bar", "foo", ".bar.txt")
	foo2bar := &Foo2Barer{
		InnerProc: innerFoo2Bar,
		InFoo:     innerFoo2Bar.InPorts["foo"],
		OutBar:    innerFoo2Bar.OutPorts["bar"],
	}
	foo2bar.InnerProc.CustomExecute = func(task *sci.ShellTask) {
		task.OutTargets["bar"].WriteTempFile(bytes.Replace(task.InTargets["foo"].Read(), []byte("foo"), []byte("bar"), 1))
	}
	return foo2bar
}

func (p *Foo2Barer) Run() {
	p.InnerProc.InPorts["foo"] = p.InFoo
	p.InnerProc.OutPorts["bar"] = p.OutBar
	p.InnerProc.Run()
}
