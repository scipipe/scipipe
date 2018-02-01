## What are re-usable components

With re-usable components, we mean components that can be stored in a Go
package and imported and used later.

In order for components in such a library to be easy to use, the ports need
to be static methods bound to the process struct, rather than just stored by
a string ID in a generic port map, like the `In()` and `Out()` methods on
`Process` processes. This is so that the methods can show up in the
auto-completion / intellisense function in code editors removing the need to
look up the name of the ports manually in the library code all the time.

## How to create re-usable components in SciPipe

Process processes created with the `scipipe.NewProc()` command, can be turned
into such "re-usable" component by using a wrapping strategy, that is
demonstrated in an [example on GitHub](https://github.com/scipipe/scipipe/blob/master/examples/wrapper_procs/wrap.go).

The idea is to create a new struct type for the re-usable component, and
then, in the factory method for the process, create an "inner" process of
type Process, using `NewProc()` as in the normal case, embedding that in the
outer struct and then adding statically defined accessor methods for each of
the ports in the inner process, with a similar name. So, if the inner process
has an outport named "foo", you would define an accessor method named
`myproc.OutFoo()` that returns this port from the inner process.

Let's look at a code example of how this works, by creating a process that just
writes "hi" to a file:

```go
type HiWriter struct {
    // Embedd a Process struct
	*sci.Process
}

func NewHiWriter() *HiWriter {
    // Initialize a normal "Process" to use as an "inner" process
	innerHiWriter := sci.NewProc("hiwriter", "echo hi > {o:hifile}")
	innerHiWriter.SetPathStatic("hifile", "hi.txt")

    // Create a new HiWriter process with the inner process embedded into it
	return &HiWriter{innerHiWriter}
}

// OutHiFile provides a static version of the "hifile" port in the inner
// (embedded) process
func (p *HiWriter) OutHiFile() *sci.OutPort {
    // Return the inner process' port named "hifile"
    return p.Out("hifile")
}
```

## See also

- [A full, working, workflow example using this trategy](https://github.com/scipipe/scipipe/blob/master/examples/wrapper_procs/wrap.go)
