package main

import (
	"bytes"
	. "github.com/scipipe/scipipe"
)

func main() {
	wf := NewWorkflow("FuncHookWf")

	foo := NewFooer("foo")
	wf.Add(foo)

	f2b := NewFoo2Barer("f2b")
	wf.Add(f2b)

	snk := NewSink("snk")
	wf.SetDriver(snk)

	foo.OutFoo().Connect(f2b.InFoo())
	snk.Connect(f2b.OutBar())

	wf.Run()
}

// ------------------------------------------------------------------------
// Components
// ------------------------------------------------------------------------

// Fooer

type Fooer struct {
	*SciProcess
	name string
}

func NewFooer(name string) *Fooer {
	// Initiate task from a "shell like" pattern, though here we
	// just specify the out-port foo
	innerFoo := NewFromShell("fooer", "{o:foo}")
	// Set the output formatter to a static string
	innerFoo.SetPathStatic("foo", "foo.txt")
	// Create the custom execute function, with pure Go code
	innerFoo.CustomExecute = func(task *SciTask) {
		task.OutTargets["foo"].WriteTempFile([]byte("foo\n"))
	}
	// Connect the ports of the outer task to the inner, generic one
	fooer := &Fooer{
		innerFoo,
		name,
	}
	return fooer
}

func (p *Fooer) OutFoo() *FilePort { return p.Out("foo") }

// Foo2Barer

type Foo2Barer struct {
	*SciProcess
	name string
}

func NewFoo2Barer(name string) *Foo2Barer {
	// Initiate task from a "shell like" pattern, though here we
	// just specify the in-port foo and the out-port bar
	innerProc := NewFromShell("foo2bar", "{i:foo}{o:bar}")
	// Set the output formatter to extend the path on the "bar"" in-port
	innerProc.SetPathExtend("foo", "bar", ".bar.txt")
	// Create the custom execute function, with pure Go code
	innerProc.CustomExecute = func(task *SciTask) {
		task.OutTargets["bar"].WriteTempFile(bytes.Replace(task.InTargets["foo"].Read(), []byte("foo"), []byte("bar"), 1))
	}

	// Connect the ports of the outer task to the inner, generic one
	return &Foo2Barer{
		innerProc,
		name,
	}
}

func (p *Foo2Barer) InFoo() *FilePort  { return p.In("foo") }
func (p *Foo2Barer) OutBar() *FilePort { return p.Out("bar") }
