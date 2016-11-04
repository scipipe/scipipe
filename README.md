# SciPipe

[![Build Status](https://travis-ci.org/scipipe/scipipe.svg?branch=master)](https://travis-ci.org/scipipe/scipipe)
[![GratiPay amount](http://img.shields.io/gratipay/samuell.svg)](https://gratipay.com/samuell)

SciPipe is an experimental library for writing [scientific Workflows](https://en.wikipedia.org/wiki/Scientific_workflow_system) in vanilla [Go(lang)](http://golang.org).
The architecture of SciPipe is based on an [flow-based programming](https://en.wikipedia.org/wiki/Flow-based_programming) like pattern in pure Go as presented in
[this](http://blog.gopheracademy.com/composable-pipelines-pattern) and [this](https://blog.gopheracademy.com/advent-2015/composable-pipelines-improvements/) Gopher Academy blog posts.

**UPDATE Nov 4, 2016:** See also [a poster on SciPipe](http://dx.doi.org/10.13140/RG.2.2.34414.61760), presented at the [e-Science Academy in Lund, Sweden, on Oct 12-13 2016](essenceofescience.se/event/swedish-e-science-academy-2016-2/).

**UPDATE June 23, 2016:** See also [slides from a recent presentation of SciPipe for use in a Bioinformatics setting](http://www.slideshare.net/SamuelLampa/scipipe-a-lightweight-workflow-library-inspired-by-flowbased-programming).

## An example workflow

Before going into details, let's look at a toy-example workflow, to get a feel
for what writing workflows with SciPipe looks like:

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

... and to see how we would run this code, let's assume we put this code in a
file `myfirstworkflow.go` and run it. Then it can look like this:

```bash
[samuel test]$ go run myfirstworkflow.go
AUDIT   2016/06/09 17:17:41 Task:foowriter    Executing command: echo 'foo' > foo.txt.tmp
AUDIT   2016/06/09 17:17:41 Task:foo2bar      Executing command: sed 's/foo/bar/g' foo.txt > foo.txt.bar.tmp
```

As you see, it displays all the shell commands it has executed based on the defined workflow.

## Benefits

Some benefits of SciPipe that are not always available in other scientific workflow systems:

- **Easy-to-grasp behaviour:** Data flowing through a network.
- **Parallel:** Apart from the inherent pipeline parallelism, SciPipe processes also spawn multiple parallel tasks when the same process has multiple inputs.
- **Concurrent:** Each process runs in an own light-weight thread, and is not blocked by
  operations in other processes, except when waiting for inputs from upstream processes.
- **Inherently simple:** Uses Go's concurrency primitives (go-routines and channels)
  to create an "implicit" scheduler, which means very little additional infrastructure code.
  This means that the code is easy to modify and extend.
- **Resource efficient:** You can choose to stream selected outputs via Unix FIFO files, to avoid temporary storage.
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
- **Notebookeable:** Works well in [Jupyter notebooks](http://jupyter.org), using the [gophernotes kernel](https://github.com/gopherds/gophernotes) (Thx [@dwhitena](https://twitter.com/dwhitena) for creating this!)

## Known limitations

- There is not yet a really comprehensive audit log generation. It is being worked on currently.
- There is not yet support for the [Common Workflow Language](http://common-workflow-language.github.io), but that is also something that we plan to support in the future.

## Connection to flow-based programming

From Flow-based programming, SciPipe uses the ideas of separate network (workflow dependency graph)
definition, named in- and out-ports, sub-networks/sub-workflows and bounded buffers (already available
in Go's channels) to make writing workflows as easy as possible.

In addition to that it adds convenience factory methods such as `scipipe.NewFromShell()` which creates ad hoc processes
on the fly based on a shell command pattern, where  inputs, outputs and parameters are defined in-line
in the shell command with a syntax of `{i:INPORT_NAME}` for inports, and `{o:OUTPORT_NAME}` for outports
and `{p:PARAM_NAME}` for parameters.

## Getting started: Install

1. Install Go by following instructions on [this page](https://golang.org/doc/install).
  - I typically install to a custom location (`~/go` for the go tools, and `~/code/go` for my own go-projects).
  - If you want to install (which means, untar the go tarball) to `~/go` just like me, you should put the following in your `~/.bashrc` file:
  
  ```bash
  # Go stuff
  export GOROOT=~/go
  export GOPATH=~/code/go
  export PATH=$GOROOT/bin:$PATH
  export PATH=$GOPATH/bin:$PATH
  ```
  
2. Then, install scipipe:
  
  ```bash
  go get github.com/scipipe/scipipe/...
  ```
**N.B:** Don't miss the `...`, or you won't get the `scipipe` helper tool.
  
3. Now, you should be able to write code like in the example below, in files ending with `.go`.
  0. The easiest way to get started is to let the scipipe tool generate a starting point for you:

  ```bash
  scipipe new myfirstworkflow.go
  ```

  ... which you can then edit to your liking.

4. To run a `.go` file, use `go run`:
  
  ```bash
  go run myfirstworkflow.go
  ```

## Writing workflows with SciPipe

Let's now go through the code example further above on the page, and look closer into how SciPipe works.

### Initializing processes

```go
foo := sp.NewFromShell("foowriter", "echo 'foo' > {o:out}")
f2b := sp.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
snk := sp.NewSink() // Will just receive file targets, doing nothing
```

For these inports and outports, channels for sending and receiving FileTargets are automatically
created and put in a hashmap added as a struct field of the process, named `In` and `Out` repectively,
Eash channel is added to the hashmap with its inport/outport name as key in the hashmap,
so that the channel can be retrieved from the hashmap using the in/outport name.

### Connecting processes into a network

Connecting outports of one process to the inport of another process is then
done with the `Connect` method available on each port object. Sink objects have
a `Connect` method too:

```go
f2b.In["foo"].Connect(foo.Out["foo"])
snk.Connect(f2b.Out["bar"])
```

(Note that the sink has just one inport, as a static struct field).

### Formatting output file paths

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

### Formatting output file paths: A nicer way

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

### Running the pipeline

So, the final part probably explains itself, but the pipeline runner component
is a very simple one that will start each component except the last one in a
separate go-routine, while the last process will be run in the main go-routine,
so as to block until the pipeline has finished.

```go
pl := sp.NewPipelineRunner()
pl.AddProcesses(foo, f2b, snk)
pl.Run()
```
### Summary

So with this, we have done everything needed to set up a file-based batch workflow system.

In summary, what we did, was to:

- Specify process dependencies by wiring outputs of the upstream processes to inports in downstream processes.
- For each outport, provide a function that will compute a suitable file name for the new file.

For more examples, see the [examples folder](https://github.com/scipipe/scipipe/tree/master/examples).

## Related tools

Find below a few tools that are more or less similar to SciPipe that are worth worth checking out before
deciding on what tool fits you best (in approximate order of similarity to SciPipe):

- [NextFlow](http://nextflow.io)
- [Luigi](https://github.com/spotify/luigi)/[SciLuigi](https://github.com/samuell/sciluigi)
- [BPipe](https://code.google.com/p/bpipe/)
- [SnakeMake](https://bitbucket.org/johanneskoester/snakemake)
- [Cuneiform](https://github.com/joergen7/cuneiform)

## Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
- This work is financed by faculty grants and other financing for Jarl Wikberg's [Pharmaceutical Bioinformatics group](http://www.farmbio.uu.se/forskning/researchgroups/pb/) of Dept. of
  Pharmaceutical Biosciences at Uppsala University, and to a smaller part also by [Swedish Research Council](http://vr.se) through the Swedish [Bioinformatics Infastructure for Life Sciences in Sweden](http://bils.se).
- Supervisor for the project is [Ola Spjuth](http://www.farmbio.uu.se/research/researchgroups/pb/olaspjuth).
- Big thanks to [Egon Elbre](http://twitter.com/egonelbre) for very helpful input on the design of the internals of the pipeline, and processes, which simplified the implementation a lot.
