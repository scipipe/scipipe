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
	InnerProcess *sci.SciProcess
	OutFoo       chan *sci.FileTarget
}

func NewFooer() *Fooer {
	// Initiate task from a "shell like" pattern, though here we
	// just specify the out-port foo
	innerFoo := sci.Shell("{o:foo}")
	// Set the output formatter to a static string
	innerFoo.SetPathFormatterString("foo", "foo.txt")
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
	InnerProcess := sci.Shell("{i:foo}{o:bar}")
	// Set the output formatter to extend the path on the "bar"" in-port
	InnerProcess.SetPathFormatterExtend("bar", "foo", ".bar.txt")
	// Connect the ports of the outer task to the inner, generic one
	foo2bar := &Foo2Barer{
		InnerProcess: InnerProcess,
		InFoo:        InnerProcess.InPorts["foo"],
		OutBar:       InnerProcess.OutPorts["bar"],
	}
	// Create the custom execute function, with pure Go code
	foo2bar.InnerProcess.CustomExecute = func(task *sci.SciTask) {
		task.OutTargets["bar"].WriteTempFile(bytes.Replace(task.InTargets["foo"].Read(), []byte("foo"), []byte("bar"), 1))
	}
	return foo2bar
}

func (p *Foo2Barer) Run() {
	// Connect inner ports to outer ones again, in order to update
	// connectivity after the workflow wiring has taken place.
	p.InnerProcess.InPorts["foo"] = p.InFoo
	p.InnerProcess.OutPorts["bar"] = p.OutBar
	// Run the inner process
	p.InnerProcess.Run()
}
