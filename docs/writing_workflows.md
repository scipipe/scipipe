# Writing Workflows - An Overview

In order to give an overview of how to write workflows in SciPipe, let's look
at the example workflow used on the front page again:

```go
package main

import (
    // Import SciPipe into the main namespace (generally frowned upon but could
    // be argued to be reasonable for short-lived workflow scripts like this)
    . "github.com/scipipe/scipipe"
)

func main() {
    // Init workflow with a name, and a number for max concurrent tasks, so we
    // don't overbook our CPU (it is recommended to set it to the number of CPU
    // cores of your computer)
    wf := NewWorkflow("hello_world", 4)

    // Initialize processes and set output file paths
    hello := wf.NewProc("hello", "echo 'Hello ' > {o:out}")
    hello.SetPathStatic("out", "hello.txt")

    world := wf.NewProc("world", "echo $(cat {i:in}) World >> {o:out}")
    world.SetPathReplace("in", "out", ".txt", "_world.txt")

    // Connect network
    world.In("in").Connect(hello.Out("out"))
    wf.ConnectLast(world.Out("out"))

    // Run workflow
    wf.Run()
}
```

Now let's go through the code example in some detail, to see what we are
actually doing.

## Initializing processes

```go
// Initialize processes from shell command patterns
hello := sp.NewProc("hello", "echo 'Hello ' > {o:out}")
world := sp.NewProc("world", "echo $(cat {i:in}) World >> {o:out}")
```

Here we are initializing two new processes, both of them based on a shell
command, using the `scipipe.NewProc()` function, which takes a processname, and
a shell command pattern as input.

### The shell command pattern

The shell command patterns, in this case `echo 'Hello ' > {o:out}` and
`echo $(cat {i:in}) World >> {o:out}`, are basically normal bash
shell commands, with the addition of "placeholders" for input and output
filenames.

Input filename placeholders are on the form `{i:INPORT-NAME}` and the output
filename placeholders are similarly of the form `{o:OUTPORT-NAME}`.  These
placeholders will be replaced with actual filenames when the command is
executed later. The reason that it a port-name is used to name them, is that
files will be queued on the channel connecting to the port, and for each set of
files on in-ports, a command will be created and executed whereafter new files
will be pulled in on the out-ports, and so on.

## Formatting output file paths

Now we need to provide some way for scipipe to figure out a suitable file name
for each of the files propagating through the "network" of processes.  This can
be done using special convenience methods on the processes, starting with
`SetPath...`. There are a few variants, of which two of them are shown here.


```go
// Configure output file path formatters for the processes created above
hello.SetPathStatic("out", "hello.txt")
world.SetPathReplace("in", "out", ".txt", "_world.txt")
```

`SetPathStatic` just takes an out-port name and a static file name to use, and
is suitable for processes which produce only one single output for a whole
workflow run.

`SetPathReplace` is slightly more advanced: It takes an in-port name, and
out-port name, and then a search-pattern in the input-filename, and a
replace-pattern for the output filename.  With the example above, our input
file named `hello.txt` will be converted into `hello_world.txt` by this path
pattern.

## Even more control over file formatting

We can actually get even more control over how file names are produced than
this, by manually supplying each process with an anonymous function that
returns file paths given a `scipipe.SciTask` object, which will be produced for
each command execution.

In order to implement the same path patterns as above, using this method, we
would write like this:

```go
// Configure output file path formatters for the processes created above
hello.SetPathCustom("out", func(t *sp.SciTask) string {
return "hello.txt"
})
world.SetPathCustom("out", func(t *sp.SciTask) string {
return strings.Replace(t.InTargets["in"].GetPath(), ".txt", "_world.txt", -1)
})
```

As you can see, this is a much more complicated way to format paths, but it can
be useful for example when needing to incorporate parameter values into file
names.

## Connecting processes into a network

Finally we need to define the data dependencies between our processes.  We do
this by connecting the outports of one process to the inport of another
process, using the `Connect` method available on each port object. We also need
to connect the final out-port of the pipeline to the workflow, so that the
workflow can pull on this port (technically pulling on a Go channel), in order
to drive the workflow.

```go
// Connect network
world.In("in").Connect(helloWriter.Out("out"))
wf.ConnectLast(world.Out("out"))
```

Note: If your "last" process does not have any outputs, you can instead set it
as the driver process of the workflow, which will replace the default driver
process which is of type [Sink](https://godoc.org/github.com/scipipe/scipipe#Sink).
So, for example, given that our `world` process did not have an output, but
maybe for example just wrote something to a web service, without producing any
files, we could do:

```go
wf.SetDriver(world)
```

## Running the pipeline

So, the final part probably explains itself, but the workflow component is a
relatively simple one that will start each component in a separate go-routine.

For technical reasons, one final process has to be run in the main go-routine
(that where the program's `main()` function runs), but as long as you have used
the `wf.ConnectLast()` method to connect the final output in your workflow as
mentioned above, you don't need to think about this, as the workflow will then
use an in-built [sink](https://godoc.org/github.com/scipipe/scipipe#Sink)
process for this purpose. Only if you need to customize things heavily, you
might want to change this.

```go
wf.Run()
```
## Summary

So with this, we have done everything needed to set up a file-based batch workflow system.

In summary, what we did, was to:

1. Initialize processes
2. For each out-port, define a file-naming strategy
3. Specify dependencies by connecting out- and in-ports
4. Run the pipeline

This actually turns out to be a fixed set of components that always need to be
included when writing workflows, so it might be good to keep them in mind and
memorize these steps, if needed.

For more examples, see the [examples folder](https://github.com/scipipe/scipipe/tree/master/examples)
in the GitHub repository.
