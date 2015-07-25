# SciPipe

An experimental library for writing Scientific (Batch) Workflows in vanilla [Go(lang)](http://golang.org),
based on an idea for a flow-based like pattern in pure Go, as presented in
[this Gopher Academy blog post](http://blog.gopheracademy.com/composable-pipelines-pattern).

From Flow-based programming, SciPipe uses the ideas of separate network (workflow dependency graph)
definition, named in- and out-ports, sub-networks/sub-workflows and bounded buffers (already available
in Go's channels) to make writing workflows as easy as possible.

In addition to that it adds convenience factory methods such as `sci.Sh()` which creates ad hoc tasks
on the fly based on a shell command pattern, where  inputs, outputs and parameters are defined in-line
in the shell command with a syntax of `{i:INPORT_NAME}` for inports, and `{o:OUTPORT_NAME}` for outports
and `{p:PARAM_NAME}` for parameters.

## Example: Creating two example tasks:

Let's look at a toy-example workflow. First the full version:

```go
package main

import (
	sci "github.com/samuell/scipipe"
)

func main() {
	// Initialize tasks
	fw := sci.Sh("echo 'foo' > {o:out}")
	f2b := sci.Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	snk := sci.NewSink() // Will just receive file targets, doing nothing

	// Add output file path formatters
	fw.OutPathFuncs["out"] = func() string {
		// Just a static one in this case (not using incoming file paths)
		return "foo.txt"
	}
	f2b.OutPathFuncs["bar"] = func() string {
		// Here, we instead re-use the file name of the task we depend
		// on (which we get on the 'foo' inport), and just
		// pad '.bar' at the end:
		return f2b.GetInPath("foo") + ".bar"
	}

	// Connect network
	f2b.InPorts["foo"] = fw.OutPorts["out"]
	snk.In = f2b.OutPorts["bar"]

	// Add to a pipeline and run
	pl := sci.NewPipeline()
	pl.AddTasks(fw, f2b, snk)
	pl.Run()
}
```

And to see what it does, let's put the code in a file `test.go` and run it:

```bash
[samuell test]$ go run test.go 
AUDIT: 2015/07/25 17:08:48 Starting task: echo 'foo' > foo.txt
AUDIT: 2015/07/25 17:08:48 Finished task: echo 'foo' > foo.txt
AUDIT: 2015/07/25 17:08:48 Starting task: sed 's/foo/bar/g' foo.txt > foo.txt.bar
AUDIT: 2015/07/25 17:08:48 Finished task: sed 's/foo/bar/g' foo.txt > foo.txt.bar
```

Now, let's go through the code above in more detail, part by part:

### Initializing tasks

```go
fw := sci.Sh("echo 'foo' > {o:out}")
f2b := sci.Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
snk := sci.NewSink() // Will just receive file targets, doing nothing
```

For these inports and outports, channels for sending and receiving FileTargets are automatically
created and put in a hashmap added as a struct field of the task, named `InPorts` and `OutPorts` repectively,
Eash channel is added to the hashmap with its inport/outport name as key in the hashmap,
so that the channel can be retrieved from the hashmap using the in/outport name.

### Connecting tasks into a network

Connecting outports of one task to the inport of another task is then done by assigning the
respective channels to the corresponding places in the hashmap:

```go
f2b.InPorts["foo"] = fw.OutPorts["out"]
snk.In = f2b.OutPorts["bar"]
```

(Note that the sink has just one inport, as a static struct field).

### Formatting output file paths

The only thing remaining after this, is to provide some way for the program to figure out a
suitable file name for each of the files propagating through this little "network" of tasks.
This is done by adding a closure (function) to another special hashmap, again keyed by
the names of the outports of the tasks. So, to define the output filenames of the two tasks
above, we would add:

```go
fw.OutPathFuncs["out"] = func() string {
	// Just statically create a file named foo.txt
	return "foo.txt"
}
f2b.OutPathFuncs["bar"] = func() string {
	// Here, we instead re-use the file name of the task we depend
	// on (which we get on the 'foo' inport), and just
	// pad '.bar' at the end:
	return f2b.GetInPath("foo") + ".bar"
}
```

### Running the pipeline

So, the final part probably explains itself, but the pipeline component is a very simple one
that will start each component except the last one in a separate go-routine, while the last
task will be run in the main go-routine, so as to block until the pipeline has finished.

```go
pl := sci.NewPipeline()
pl.AddTasks(fw, f2b, snk)
pl.Run()
```


### Summary

So with this, we have done everything needed to set up a file-based batch workflow system.

In summary, what we did, was to:

- Specify task dependencies by wiring outputs of the upstream tasks to inports in downstream tasks.
- For each outport, provide a function that will compute a suitable file name for the new file.

For more examples, see the [examples folder](https://github.com/samuell/scipipe/tree/master/examples).

## Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
- This work is financed by faculty grants and other financing for Jarl Wikberg's [Pharmaceutical Bioinformatics group](http://www.farmbio.uu.se/forskning/researchgroups/pb/) of Dept. of
  Pharmaceutical Biosciences at Uppsala University. Main supervisor for the project is [Ola Spjuth](http://www.farmbio.uu.se/research/researchgroups/pb/olaspjuth).
- Big thanks to [Egon Elbre](http://twitter.com/egonelbre) for very helpful input on the design of the internals of the pipeline, and tasks, which simplified the implementation a lot.
