package main

import (
	"bytes"

	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogInfo()

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
	// Initiate task from a "shell like" pattern, though here we
	// just specify the out-port foo
	innerFoo := sci.Shell("{o:foo}")
	// Set the output formatter to a static string
	innerFoo.SetPathFormatterString("foo", "foo.txt")
	// Create the custom execute function, with pure Go code
	innerFoo.CustomExecute = func(task *sci.ShellTask) {
		task.OutTargets["foo"].WriteTempFile([]byte("foo\n"))
	}
	// Connect the ports of the outer task to the inner, generic one
	fooer := &Fooer{
		InnerProc: innerFoo,
		OutFoo:    innerFoo.OutPorts["foo"],
	}
	return fooer
}

func (p *Fooer) Run() {
	// Connect inner ports to outer ones again, in order to update
	// connectivity after the workflow wiring has taken place.
	p.InnerProc.OutPorts["foo"] = p.OutFoo
	// Run the inner process
	p.InnerProc.Run()
}

// Foo2Barer

type Foo2Barer struct {
	InnerProc *sci.ShellProcess
	InFoo     chan *sci.FileTarget
	OutBar    chan *sci.FileTarget
}

func NewFoo2Barer() *Foo2Barer {
	// Initiate task from a "shell like" pattern, though here we
	// just specify the in-port foo and the out-port bar
	innerFoo2Bar := sci.Shell("{i:foo}{o:bar}")
	// Set the output formatter to extend the path on the "bar"" in-port
	innerFoo2Bar.SetPathFormatterExtend("bar", "foo", ".bar.txt")
	// Connect the ports of the outer task to the inner, generic one
	foo2bar := &Foo2Barer{
		InnerProc: innerFoo2Bar,
		InFoo:     innerFoo2Bar.InPorts["foo"],
		OutBar:    innerFoo2Bar.OutPorts["bar"],
	}
	// Create the custom execute function, with pure Go code
	foo2bar.InnerProc.CustomExecute = func(task *sci.ShellTask) {
		task.OutTargets["bar"].WriteTempFile(bytes.Replace(task.InTargets["foo"].Read(), []byte("foo"), []byte("bar"), 1))
	}
	return foo2bar
}

func (p *Foo2Barer) Run() {
	// Connect inner ports to outer ones again, in order to update
	// connectivity after the workflow wiring has taken place.
	p.InnerProc.InPorts["foo"] = p.InFoo
	p.InnerProc.OutPorts["bar"] = p.OutBar
	// Run the inner process
	p.InnerProc.Run()
}
