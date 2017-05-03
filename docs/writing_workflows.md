# Writing workflows with SciPipe

## An example workflow

Before going into details about how to write SciPipe workflows, let's look at
the example workflow used on the front page, and use it as an example when we
discuss the SciPipe syntax further below:

```go
package main

import (
    sp "github.com/scipipe/scipipe"
)

func main() {
    // Initialize processes
    foo := sp.NewFromShell("foowriter", "echo 'foo' > {o:foo}")
    f2b := sp.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
    snk := sp.NewSink() // Will just receive file targets, doing nothing

    // Add output file path formatters for the components created above
    foo.SetPathStatic("foo", "foo.txt")
    f2b.SetPathExtend("foo", "bar", ".bar")

    // Connect network
    f2b.In["foo"].Connect(foo.Out["foo"])
    snk.Connect(f2b.Out["bar"])

    // Add to a pipeline runner and run
    pl := sp.NewPipelineRunner()
    pl.AddProcesses(foo, f2b, snk)
    pl.Run()
}
```

Let us now go through the code example step by step, and describe in more
detail what we are doing.

## Initializing processes

```go
foo := sp.NewFromShell("foowriter", "echo 'foo' > {o:out}")
f2b := sp.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
snk := sp.NewSink() // Will just receive file targets, doing nothing
```

For these inports and outports, channels for sending and receiving FileTargets are automatically
created and put in a hashmap added as a struct field of the process, named `In` and `Out` repectively,
Eash channel is added to the hashmap with its inport/outport name as key in the hashmap,
so that the channel can be retrieved from the hashmap using the in/outport name.

## Connecting processes into a network

Connecting outports of one process to the inport of another process is then
done with the `Connect` method available on each port object. Sink objects have
a `Connect` method too:

```go
f2b.In["foo"].Connect(foo.Out["foo"])
snk.Connect(f2b.Out["bar"])
```

(Note that the sink has just one inport, as a static struct field).

## Formatting output file paths

The only thing remaining after this, is to provide some way for the program to figure out a
suitable file name for each of the files propagating through this little "network" of processes.
This is done by adding a closure (function) to another special hashmap, again keyed by
the names of the outports of the processes. So, to define the output filenames of the two processes
above, we would add:

```go
foo.PathFormatters["foo"] = func(t *sp.SciTask) string {
	// Just statically create a file named foo.txt
	return "foo.txt"
}
f2b.PathFormatters["bar"] = func(t *sp.SciTask) string {
	// Here, we instead re-use the file name of the process we depend
	// on (which we get on the 'foo' inport), and just
	// pad '.bar' at the end:
	return f2b.GetInPath("foo") + ".bar"
}
```

## Formatting output file paths: A nicer way

Now, the above way of defining path formats is a bit verbose, isn't it?
Luckily, there's a shorter way, by using convenience methods for doing the same
thing. So, the above two path formats can also be defined like so, with the exact same result:

```go
// Create a static file name for the out-port 'foo':
foo.SetPathStatic("foo", "foo.txt")

// For out-port 'bar', extend the file names of files on in-port 'foo', with
// the suffix '.bar':
f2b.SetPathExtend("foo", "bar", ".bar")
```

## Running the pipeline

So, the final part probably explains itself, but the pipeline runner component
is a very simple one that will start each component except the last one in a
separate go-routine, while the last process will be run in the main go-routine,
so as to block until the pipeline has finished.

```go
pl := sp.NewPipelineRunner()
pl.AddProcesses(foo, f2b, snk)
pl.Run()
```
## Summary

So with this, we have done everything needed to set up a file-based batch workflow system.

In summary, what we did, was to:

- Specify process dependencies by wiring outputs of the upstream processes to inports in downstream processes.
- For each outport, provide a function that will compute a suitable file name for the new file.

For more examples, see the [examples folder](https://github.com/scipipe/scipipe/tree/master/examples).
