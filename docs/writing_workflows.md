# Writing Workflows - An Overview

In order to give an overview of how to write workflows in SciPipe, let's look
at the example workflow used on the front page again:

```go
package main

import (
    // Import SciPipe, aliased to 'sp' for brevity
    sp "github.com/scipipe/scipipe"
)

func main() {
    // Initialize processes from shell command patterns
    helloWriter := sp.NewProc("helloWriter", "echo 'Hello ' > {o:hellofile}")
    worldAppender := sp.NewProc("worldAppender", "echo $(cat {i:infile}) World >> {o:worldfile}")
    // Create a sink, that will just receive the final outputs
    sink := sp.NewSink("sink")

    // Configure output file path formatters for the processes created above
    helloWriter.SetPathStatic("hellofile", "hello.txt")
    worldAppender.SetPathReplace("infile", "worldfile", ".txt", "_world.txt")

    // Connect network
    worldAppender.In("infile").Connect(helloWriter.Out("hellofile"))
    sink.Connect(worldAppender.Out("worldfile"))

    // Create a pipeline runner, add processes, and run
    wf := sp.NewWorkflow("example_workflow")
    wf.Add(helloWriter, worldAppender)
    wf.SetDriver(sink)
    wf.Run()
}
```

Now let's go through the code example in some detail, to see what we are
actually doing.

## Initializing processes

```go
// Initialize processes from shell command patterns
helloWriter := sp.NewProc("helloWriter", "echo 'Hello ' > {o:hellofile}")
worldAppender := sp.NewProc("worldAppender", "echo $(cat {i:infile}) World >> {o:worldfile}")
// Create a sink, that will just receive the final outputs
sink := sp.NewSink("sink")
```

Here we are initializing three new processes, two of them based on a shell
command, and one "sink", which will just receive inputs adn nothing more.

The two first processes are created using the `scipipe.NewProc()`
function, which takes a processname, and a shell command pattern as input.

### The shell command pattern

The shell command patterns, in this case `echo 'Hello ' > {o:hellofile}` and
`echo $(cat {i:infile}) World >> {o:worldfile}`, are basically normal bash
shell commands, with the addition of "placeholders" for input and output
filenames.

Input filename placeholders are on the form `{i:INPORT-NAME}` and the output
filename placeholders are similarly of the form `{o:OUTPORT-NAME}`.  These
placeholders will be replaced with actual filenames when the command is
executed later. The reason that it a port-name is used to name them, is that
files will be queued on the channel connecting to the port, and for each set of
files on in-ports, a command will be created and executed whereafter new files
will be pulled in on the out-ports, and so on.

### The sink

The sink is needed in cases where the workflow ends with a process that is not
an explicit endpoint without out-ports, such as a "printer" processes or
similar, but instead has out-ports that need to be connected. Then the sink can
be used to receive from these out-ports so that the data packets on the
out-ports don't get stuck and clog the workflow.

For these inports and outports, channels for sending and receiving FileTargets
are automatically created and put in a hashmap added as a struct field of the
process, named `In` and `Out` repectively, Eash channel is added to the hashmap
with its inport/outport name as key in the hashmap, so that the channel can be
retrieved from the hashmap using the in/outport name.

## Formatting output file paths

Now we need to provide some way for scipipe to figure out a suitable file name
for each of the files propagating through the "network" of processes.  This can
be done using special convenience methods on the processes, starting with
`SetPath...`. There are a few variants, of which two of them are shown here.


```go
// Configure output file path formatters for the processes created above
helloWriter.SetPathStatic("hellofile", "hello.txt")
worldAppender.SetPathReplace("infile", "worldfile", ".txt", "_world.txt")
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
helloWriter.PathFormatters["hellofile"] = func(t *sp.SciTask) string {
return "hello.txt"
}
worldAppender.PathFormatters["worldfile"] = func(t *sp.SciTask) string {
return strings.Replace(t.InTargets["infile"].GetPath(), ".txt", "_world.txt", -1)
}
```

As you can see, this is a much more complicated way to format paths, but it can
be useful for example when needing to incorporate parameter values into file
names.

## Connecting processes into a network

Finally we need to define the data dependencies between our processes.  We do
this by connecting the outports of one process to the inport of another
process, using the `Connect` method available on each port object. Sink objects
have a `Connect` method too, which take an out-port of an upstream process:

```go
// Connect network
worldAppender.In("infile").Connect(helloWriter.Out("hellofile"))
sink.Connect(worldAppender.Out("worldfile"))
```

(Note that the sink has the `Connect` method bound directly to itself, without
any port).

## Running the pipeline

So, the final part probably explains itself, but the workflow component is a
relatively simple one that will start each component except the last one in a
separate go-routine, except for the one set as "driver" (often a simple "sink"
process), which will be run in the main go-routine, so as to block until the
pipeline has finished.

```go
// Create a pipeline runner, add processes, and run
wf := sp.NewWorkflow()
wf.Add(helloWriter, worldAppender)
wf.SetDriver(sink)
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
