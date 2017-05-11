*Beware: Somewhat technical topic, probably best suited for power-users*

## What are re-usable components

With re-usable components, we mean components that can be stored in a Go
package, and imported and used later.

In order for components in such a library to be easy to use, the ports need to
be static fields on the process, rather than just stored by a string ID in a
generic port map, like the `In` and `Out` fields on `SciProcess` processes.
This is so that the fields can show up in the auto-completion / intellisense
function in code editors, so that one does not need to look up the name of the
ports manually in the library code all the time.

## How to create re-usable components in SciPipe

SciProcess processes created with the `scipipe.NewFromShell()` command, can be
turned into such "re-usable" component by using a wrapping strategy, that is
demonstrated in an [example on GitHub](https://github.com/scipipe/scipipe/blob/master/examples/wrapper_tasks/wrap.go).

The idea is to create a new struct type for the re-usable component, and then,
in the factory method for the process, create an "inner" process of type
SciProcess, using `NewFromShell()` as in the normal case, and then making sure
that the ports of the inner process are added to the port-fields also of the
outer, "wrapping" process.

There is only a little caveat, or trick, needed to get this to work properly:
In the Run() method, which is executed after the outer process has been
connected to other processes in a network, we need to set the ports of the
inner process to be the same as the ports of the outer process. This is because
the port object might have been changed when the workflow was connected. I.e,
the process port fields might have got assigned a port object from another
process in the workflow.

The full implementation of a process that just writes "hi" to a file, can look
like this:

```go
type HiWriter struct {
	InnerProc *sci.SciProcess
	OutHi  *sci.FilePort
}

func NewHiWriter() *HiWriter {
    // Initialize a normal "SciProcess" to use as an "inner" process
	innerHiWriter := sci.NewFromShell("hiwriter", "echo hi > {o:hifile}")
	innerHiWriter.SetPathStatic("hifile", "hi.txt")

    // Create a new HiWriter process (the outer one) with the inner process
    // added to the InnerProcess field
	return &HiWriter{
		InnerProc: innerHiWriter,
		OutHi:  sci.NewFilePort(),
	}
}

func (p *HiWriter) Run() {
    // Make sure the inner process' port object is the same as the outer one's
	p.InnerProc.Out["hifile"] = p.OutHi
	p.InnerProc.Run()
}

func (p *HiWriter) IsConnected() bool {
	return p.OutHi.IsConnected()
}
```

## See also

- [A full, working, workflow example using this trategy](https://github.com/scipipe/scipipe/blob/master/examples/wrapper_tasks/wrap.go)
