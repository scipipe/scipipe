# SciPipe

An experimental library for writing Scientific (Batch) Workflows in vanilla [Go(lang)](http://golang.org),
based on an idea for a flow-based like pattern in pure Go, as presented in
[this Gopher Academy blog post](http://blog.gopheracademy.com/composable-pipelines-pattern).

See a short demo about how it might look, working with SciPipe in the terminal:

![SciPipe Demo](http://i.imgur.com/pPPAJCx.gifv)

## Benefits

Some benefits of SciPipe, that are not always available in other systems available:

- **Easy-to-grasp behaviour:** Data flowing through a network.
- **Parallel:** Apart from the inherent pipeline parallelism, SciPipe processes also spawn multiple parallel tasks when the same process has multiple inputs.
- **Concurrent:** Each process runs in an own light-weight thread, and is not blocked by
  operations in other processes, except when waiting for inputs from upstream processes.
- **Inherently simple:** Uses Go's concurrency primitives (go-routines and channels)
  to create an "implicit" scheduler, which means very little additional infrastructure code.
  This means that the code is easy to modify and extend.
- **Flexible:** Processes that wrap command-line programs and scripts can be combined with
  processes coded directly in Golang.
- **Custom file naming:** SciPipe gives you full control over how file names are produced,
  making it easy to understand and find your way among the output files of your computations.
- **Highly Debuggable(!):** Since everything in SciPipe is plain Go(lang), you can easily use the [gdb debugger](http://golang.org/doc/gdb) (preferrably
  with the [cgdb interface](https://www.youtube.com/watch?v=OKLR6rrsBmI) for easier use) to step through your program at any detail, as well as all
  the other excellent debugging tooling for Go (See eg [delve](https://github.com/derekparker/delve) and [godebug](https://github.com/mailgun/godebug)),
  or just use `println()` statements at any place in your code. In addition, you can easily
  turn on very detailed debug output from SciPipe's execution itself, by just turning on debug-level
  logging with `scipipe.InitLogDebug()` in your `main()` method.
- **Efficient:** Workflows are compiled into static compiled code, that runs fast.
- **Portable:** Workflows can be distributed as go code to be run with the `go run` command
  or compiled into stand-alone binaries for basically any unix-like operating system.

## Known limitations

- There is not yet a really comprehensive audit log generation. It is being worked on currently.
- There is as of yet no streaming support, but we have ideas and plans for how to implement that shortly.
- There is not yet support for the [Common Workflow Language](http://common-workflow-language.github.io), but that is also something that we plan to support in the future.

## Connection to flow-based programming

From Flow-based programming, SciPipe uses the ideas of separate network (workflow dependency graph)
definition, named in- and out-ports, sub-networks/sub-workflows and bounded buffers (already available
in Go's channels) to make writing workflows as easy as possible.

In addition to that it adds convenience factory methods such as `sci.Shell()` which creates ad hoc processes
on the fly based on a shell command pattern, where  inputs, outputs and parameters are defined in-line
in the shell command with a syntax of `{i:INPORT_NAME}` for inports, and `{o:OUTPORT_NAME}` for outports
and `{p:PARAM_NAME}` for parameters.

## An example workflow

Let's look at a toy-example workflow. First the full version:

```go
package main

import (
	sp "github.com/samuell/scipipe"
)

func main() {
	// Initialize processes
	fwt := sp.Shell("echo 'foo' > {o:foo}")
	f2b := sp.Shell("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	snk := sp.NewSink() // Will just receive file targets, doing nothing

	// Add output file path formatters
	fwt.OutPathFuncs["foo"] = func() string {
		// Just a static one in this case (not using incoming file paths)
		return "foo.txt"
	}
	f2b.OutPathFuncs["bar"] = func() string {
		// Here, we instead re-use the file name of the process we depend
		// on (which we get on the 'foo' inport), and just
		// pad '.bar' at the end:
		return f2b.GetInPath("foo") + ".bar"
	}

	// Connect network
	f2b.InPorts["foo"] = fwt.OutPorts["foo"]
	snk.In = f2b.OutPorts["bar"]

	// Add to a pipeline and run
	pl := sp.NewPipeline()
	pl.AddProcs(fwt, f2b, snk)
	pl.Run()
}
```

And to see what it does, let's put the code in a file `test.go` and run it:

```bash
[samuell test]$ go run test.go
AUDIT: 2015/07/25 17:08:48 Starting process: echo 'foo' > foo.txt
AUDIT: 2015/07/25 17:08:48 Finished process: echo 'foo' > foo.txt
AUDIT: 2015/07/25 17:08:48 Starting process: sed 's/foo/bar/g' foo.txt > foo.txt.bar
AUDIT: 2015/07/25 17:08:48 Finished process: sed 's/foo/bar/g' foo.txt > foo.txt.bar
```

Now, let's go through the code above in more detail, part by part:

### Initializing processes

```go
fwt := sci.Shell("echo 'foo' > {o:out}")
f2b := sci.Shell("sed 's/foo/bar/g' {i:foo} > {o:bar}")
snk := sci.NewSink() // Will just receive file targets, doing nothing
```

For these inports and outports, channels for sending and receiving FileTargets are automatically
created and put in a hashmap added as a struct field of the process, named `InPorts` and `OutPorts` repectively,
Eash channel is added to the hashmap with its inport/outport name as key in the hashmap,
so that the channel can be retrieved from the hashmap using the in/outport name.

### Connecting processes into a network

Connecting outports of one process to the inport of another process is then done by assigning the
respective channels to the corresponding places in the hashmap:

```go
f2b.InPorts["foo"] = fwt.OutPorts["out"]
snk.In = f2b.OutPorts["bar"]
```

(Note that the sink has just one inport, as a static struct field).

### Formatting output file paths

The only thing remaining after this, is to provide some way for the program to figure out a
suitable file name for each of the files propagating through this little "network" of processes.
This is done by adding a closure (function) to another special hashmap, again keyed by
the names of the outports of the processes. So, to define the output filenames of the two processes
above, we would add:

```go
fwt.OutPathFuncs["out"] = func() string {
	// Just statically create a file named foo.txt
	return "foo.txt"
}
f2b.OutPathFuncs["bar"] = func() string {
	// Here, we instead re-use the file name of the process we depend
	// on (which we get on the 'foo' inport), and just
	// pad '.bar' at the end:
	return f2b.GetInPath("foo") + ".bar"
}
```

### Running the pipeline

So, the final part probably explains itself, but the pipeline component is a very simple one
that will start each component except the last one in a separate go-routine, while the last
process will be run in the main go-routine, so as to block until the pipeline has finished.

```go
pl := sci.NewPipeline()
pl.AddProcs(fwt, f2b, snk)
pl.Run()
```
### Summary

So with this, we have done everything needed to set up a file-based batch workflow system.

In summary, what we did, was to:

- Specify process dependencies by wiring outputs of the upstream processes to inports in downstream processes.
- For each outport, provide a function that will compute a suitable file name for the new file.

For more examples, see the [examples folder](https://github.com/samuell/scipipe/tree/master/examples).

## Related tools

Find below a few tools that are more or less similar to SciPipe that are worth worth checking out before
deciding on what tool fits you best (in approximate order of similarity to SciPipe):

- [NextFlow](http://nextflow.io)
- [Luigi](https://github.com/spotify/luigi)/[SciLuigi](https://github.com/samuell/sciluigi)
- [BPipe](https://code.google.com/p/bpipe/)
- [SnakeMake](https://bitbucket.org/johanneskoester/snakemake)

## Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
- This work is financed by faculty grants and other financing for Jarl Wikberg's [Pharmaceutical Bioinformatics group](http://www.farmbio.uu.se/forskning/researchgroups/pb/) of Dept. of
  Pharmaceutical Biosciences at Uppsala University, and to a smaller part also by [Swedish Research Council](http://vr.se) through the Swedish [Bioinformatics Infastructure for Life Sciences in Sweden](http://bils.se).
- Supervisor for the project is [Ola Spjuth](http://www.farmbio.uu.se/research/researchgroups/pb/olaspjuth).
- Big thanks to [Egon Elbre](http://twitter.com/egonelbre) for very helpful input on the design of the internals of the pipeline, and processes, which simplified the implementation a lot.
