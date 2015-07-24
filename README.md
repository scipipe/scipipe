# SciPipe

An experimental library for writing Scientific (Batch) Workflows in vanilla [Go(lang)](http://golang.org).

This is an experiment in building a scientific workflow engine in pure Go,
based on an idea for a flow-based like pattern in pure Go, as presented by the author in
[this blog post on Gopher Academy](http://blog.gopheracademy.com/composable-pipelines-pattern).

From flow-based programming, It uses the principles of separate network (workflow dependency graph)
definition, named in- and out-ports, sub-networks/sub-workflows, and bounded buffers (already available 
in Go's channels) to make writing workflows as easy as possible.

In addition to that, it adds convenience factory methods (see `sp.Sh()` below) for creating ad hoc tasks
on the fly based on a shell command pattern, where  inputs, outputs and parameters are defined in-line 
in the shell command with a syntax of `{i:INPORT_NAME}` for inports, and `{o:OUTPORT_NAME}` for outports
and `{p:PARAM_NAME}` for parameters.

## Example: Creating two example tasks:

Let's look at a toy example workflow. First the full version:

```go
// Initialize tasks
fw := sp.Sh("echo 'foo' > {o:outfile}")
f2b := sp.Sh("cat {i:foofile} | sed 's/foo/bar/g' > {o:barfile}")

// Add output file path formatters
fw.OutPathFuncs["outfile"] = func() string {
	// Just statically create a file named foo.txt
	return "foo.txt"
}
f2b.OutPathFuncs["barfile"] = func() string {
	// Here, we instead re-use the file name of the task we depend
	// on (which we get on the 'foofile' inport), and just
	// pad '.bar' at the end:
	return f2b.GetInPath("foofile") + ".bar"
}

// Connect network
f2b.InPorts["foofile"] = fw.OutPorts["outfile"]

// Add to a pipeline and run
pl := sp.NewPipeline()
pl.AddTasks(fw, f2b)
pl.Run()
```

Now, let's go through the code above in more detail, part by part:

### Initializing tasks

```go
fw := sp.Sh("echo 'foo' > {o:outfile}")
f2b := sp.Sh("cat {i:foofile} | sed 's/foo/bar/g' > {o:barfile}")

```

For these inports and outports, channels for sending and receiving FileTargets are automatically
created and put in a hashmap added as a struct field of the task, named `InPorts` and `OutPorts` repectively,
Eash channel is added to the hashmap with its inport/outport name as key in the hashmap,
so that the channel can be retrieved from the hashmap using the in/outport name.

### Connecting tasks into a network

Connecting outports of one task to the inport of another task is then done by assigning the
respective channels to the corresponding places in the hashmap.

Example: Connecting the two tasks creating above:

```go
f2b.InPorts["foofile"] = fw.OutPorts["outfile"]
```

### Formatting output file paths

The only thing remaining after this, is to provide some way for the program to figure out a
suitable file name for each of the files propagating through this little "network" of tasks.
This is done by adding a closure (function) to another special hashmap, again keyed by
the names of the outports of the tasks. So, to define the output filenames of the two tasks
above, we would add:

```go
fw.OutPathFuncs["outfile"] = func() string {
	// Just statically create a file named foo.txt
	return "foo.txt"
}
f2b.OutPathFuncs["barfile"] = func() string {
	// Here, we instead re-use the file name of the task we depend
	// on (which we get on the 'foofile' inport), and just
	// pad '.bar' at the end:
	return f2b.GetInPath("foofile") + ".bar"
}
```

### Running the pipeline

So, the final part probably explains itself, but the pipeline component is a very simple one
that will start each component except the last one in a separate go-routine, while the last
task will be run in the main go-routine, so as to block until the pipeline has finished.

```go
pl := sp.NewPipeline()
pl.AddTasks(fw, f2b)
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
