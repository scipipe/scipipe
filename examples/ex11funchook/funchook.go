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

	pl := sci.NewPipelineRunner()
	pl.AddProcs(foo, f2b, snk)
	pl.Run()
}

// ------------------------------------------------------------------------
// Components
// ------------------------------------------------------------------------

// Fooer

type Fooer struct {
	InnerProcess *sci.SciProcess
	OutFoo       chan *sci.FileTarget
}

func NewFooer() *Fooer {
	// Initiate task from a "shell like" pattern, though here we
	// just specify the out-port foo
	innerFoo := sci.Shell("fooer", "{o:foo}")
	// Set the output formatter to a static string
	innerFoo.SetPathFormatStatic("foo", "foo.txt")
	// Create the custom execute function, with pure Go code
	innerFoo.CustomExecute = func(task *sci.SciTask) {
		task.OutTargets["foo"].WriteTempFile([]byte("foo\n"))
	}
	// Connect the ports of the outer task to the inner, generic one
	fooer := &Fooer{
		InnerProcess: innerFoo,
		OutFoo:       innerFoo.OutPorts["foo"],
	}
	return fooer
}

func (p *Fooer) Run() {
	// Connect inner ports to outer ones again, in order to update
	// connectivity after the workflow wiring has taken place.
	p.InnerProcess.OutPorts["foo"] = p.OutFoo
	// Run the inner process
	p.InnerProcess.Run()
}

// Foo2Barer

type Foo2Barer struct {
	InnerProcess *sci.SciProcess
	InFoo        chan *sci.FileTarget
	OutBar       chan *sci.FileTarget
}

func NewFoo2Barer() *Foo2Barer {
	// Initiate task from a "shell like" pattern, though here we
	// just specify the in-port foo and the out-port bar
	innerProc := sci.Shell("foo2bar", "{i:foo}{o:bar}")
	// Set the output formatter to extend the path on the "bar"" in-port
	innerProc.SetPathFormatExtend("bar", "foo", ".bar.txt")
	// Create the custom execute function, with pure Go code
	innerProc.CustomExecute = func(task *sci.SciTask) {
		task.OutTargets["bar"].WriteTempFile(bytes.Replace(task.InTargets["foo"].Read(), []byte("foo"), []byte("bar"), 1))
	}

	// Connect the ports of the outer task to the inner, generic one
	return &Foo2Barer{
		InnerProcess: innerProc,
		InFoo:        innerProc.InPorts["foo"],
		OutBar:       innerProc.OutPorts["bar"],
	}
}

func (p *Foo2Barer) Run() {
	// Connect inner ports to outer ones again, in order to update
	// connectivity after the workflow wiring has taken place.
	p.InnerProcess.InPorts["foo"] = p.InFoo
	p.InnerProcess.OutPorts["bar"] = p.OutBar
	// Run the inner process
	p.InnerProcess.Run()
}
