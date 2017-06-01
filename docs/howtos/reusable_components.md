## What are re-usable components

With re-usable components, we mean components that can be stored in a Go
package and imported and used later.

In order for components in such a library to be easy to use, the ports need to
be static fields, or even better --- methods, on the process, rather than just
stored by a string ID in a generic port map, like the `In` and `Out` fields on
`SciProcess` processes.  This is so that the methods can show up in the
auto-completion / intellisense function in code editors, so you don't need to
look up the name of the ports manually in the library code all the time.

## How to create re-usable components in SciPipe

SciProcess processes created with the `scipipe.NewFromShell()` command, can be
turned into such "re-usable" component by using a wrapping strategy, that is
demonstrated in an [example on GitHub](https://github.com/scipipe/scipipe/blob/master/examples/wrapper_procs/wrap.go).

The idea is to create a new struct type for the re-usable component, and then,
in the factory method for the process, create an "inner" process of type
SciProcess, using `NewFromShell()` as in the normal case, and then adding statically defined
accessor methods for each of the ports in the inner process, with a similar name.
So, if the inner process has an outport named "foo", you would define an accessor method named `myproc.OutFoo()`
that returns this port from the inner process.

Let's look at a code example of how this works, by creating a process that just
writes "hi" to a file:

```go
type HiWriter struct {
	InnerProc *sci.SciProcess
}

func (p *HiWriter) OutHiFile() *sci.FilePort {
    // Return the inner process' port named "hifile"
    return p.InnerProc.Out("hifile")
}

func NewHiWriter() *HiWriter {
    // Initialize a normal "SciProcess" to use as an "inner" process
	innerHiWriter := sci.NewFromShell("hiwriter", "echo hi > {o:hifile}")
	innerHiWriter.SetPathStatic("hifile", "hi.txt")

    // Create a new HiWriter process (the outer one) with the inner process
    // added to the InnerProcess field
	return &HiWriter{
		InnerProc: innerHiWriter,
	}
}

func (p *HiWriter) Run() {
	p.InnerProc.Run()
}

func (p *HiWriter) IsConnected() bool {
	return p.OutHiFile().IsConnected()
}
```

## See also

- [A full, working, workflow example using this trategy](https://github.com/scipipe/scipipe/blob/master/examples/wrapper_procs/wrap.go)
