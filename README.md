# SciPipe

[![Build Status](https://travis-ci.org/scipipe/scipipe.svg?branch=master)](https://travis-ci.org/scipipe/scipipe)

<img src="docs/images/fbp_factory.png" align="right">SciPipe is a library
for writing [scientific Workflows](https://en.wikipedia.org/wiki/Scientific_workflow_system),
in the [Go programming language](http://golang.org).

So, when you have multiple commandline applications that need to be run after
each other in a chain, where one program depends on the output of the other,
you can define these dependencies in SciPipe. SciPipe will then take care of
running the programs in the right order, at the right time, and produce audit
reports about exactly what was run.

SciPipe is especially well suited for complex dependency networks, where you
also don't always know how many outputs are produced by a particular program.
SciPipe also gives you a lot of flexibility in how your files are named.
SciPipe has many more benefits, which are listed under the [Benefits section](#benefits)
below.

## Project links

- For documentation and more detailed information about SciPipe, see [scipipe.org](http://scipipe.org)
- For reporting issues, please use the [issue tracker](https://github.com/scipipe/scipipe/issues)
- For general questions, see the [mailing list](https://groups.google.com/forum/#!forum/scipipe)

## An example workflow

Let's look at a simple toy example of a workflow, to get a feel for what
writing workflows with SciPipe looks like:

```go
package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	// Initialize processes
	fooWriter := sp.NewFromShell("foowriter", "echo 'foo' > {o:foo}")
	fooToBar := sp.NewFromShell("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	sink := sp.NewSink() // Will just receive file targets, doing nothing

	// Add output file path formatters for the components created above
	fooWriter.SetPathStatic("foo", "foo.txt")
	fooToBar.SetPathExtend("foo", "bar", ".bar")

	// Connect network
	fooToBar.In["foo"].Connect(fooWriter.Out["foo"])
	sink.Connect(fooToBar.Out["bar"])

	// Add to a pipeline runner and run
	pipeline := sp.NewPipelineRunner()
	pipeline.AddProcesses(foowriter, fooToBar, sink)
	pipeline.Run()
}
```

### Running the example workflow

Let's assume we put the code in a file named `myfirstworkflow.go` and run it.
Then it can look like this:

```bash
[samuel test]$ go run myfirstworkflow.go
AUDIT   2016/06/09 17:17:41 Task:foowriter    Executing command: echo 'foo' > foo.txt.tmp
AUDIT   2016/06/09 17:17:41 Task:foo2bar      Executing command: sed 's/foo/bar/g' foo.txt > foo.txt.bar.tmp
```

As you see, scipipe displays all the shell commands it has executed based on the defined workflow.

## Benefits

Some benefits of SciPipe that are not always available in other scientific workflow systems:

- **Easy-to-grasp behaviour:** Data flowing through a network.
- **Flexible:** Processes that wrap command-line programs and scripts can be combined with
  processes coded directly in Golang.
- **Efficient:** Workflows are compiled into static compiled code, that runs fast.
- **Portable:** Workflows can be distributed as go code to be run with the `go run` command
  or compiled into stand-alone binaries for basically any unix-like operating system.
- **Custom file naming:** SciPipe gives you full control over how file names are produced,
  making it easy to understand and find your way among the output files of your computations.
- **Inherently simple:** SciPipe uses the in-built concurrency primitives in
  the Go programming language (go-routines and channels) to create an
  "implicit" scheduler, which means very little additional infrastructure code.
  This means that the code is easy to modify and extend.
- **Supports streaming:** You can choose to stream selected outputs via Unix FIFO files, to avoid temporary storage.
- **Parallel:** SciPipe leverages both pipeline parallelism between multiple
  processes, and task parallelism when there is multiple inputs to a process,
  to make your computations complete as fast as possible, utilizing all the CPU
  cores available.
- **Notebookeable:** Works well in [Jupyter notebooks](http://jupyter.org),
  using the [gophernotes kernel](https://github.com/gopherds/gophernotes).
- **Concurrent:** Each process runs in an own light-weight thread, and is not blocked by
  operations in other processes, except when waiting for inputs from upstream processes.
- **Easy to debug:** Since everything in SciPipe is just Go code, you can
  easily use the [gdb debugger](http://golang.org/doc/gdb) (with the [cgdb
  interface](https://www.youtube.com/watch?v=OKLR6rrsBmI) for easier use) to
  step through your program at any detail, as well as all the other excellent
  debugging tooling for Go (See eg
  [delve](https://github.com/derekparker/delve) and
  [godebug](https://github.com/mailgun/godebug)), or just use `println()`
  statements at any place in your code. In addition, you can easily turn on
  detailed debug output from SciPipe's execution itself, by just turning
  on debug-level logging with `scipipe.InitLogDebug()` in your `main()` method.

## Known limitations

- There are still a number of missing good-to-have features, for workflow design. See the [issues](https://github.com/scipipe/scipipe/issues) tracker for details.
- There is not yet support for the [Common Workflow Language](http://common-workflow-language.github.io), but that is also something that we plan to support in the future.

## Connection to flow-based programming

From Flow-based programming, SciPipe uses the ideas of separate network (workflow dependency graph)
definition, named in- and out-ports, sub-networks/sub-workflows and bounded buffers (already available
in Go's channels) to make writing workflows as easy as possible.

In addition to that it adds convenience factory methods such as `scipipe.NewFromShell()` which creates ad hoc processes
on the fly based on a shell command pattern, where  inputs, outputs and parameters are defined in-line
in the shell command with a syntax of `{i:INPORT_NAME}` for inports, and `{o:OUTPORT_NAME}` for outports
and `{p:PARAM_NAME}` for parameters.

## Publications mentioning SciPipe

- See [a poster on SciPipe](http://dx.doi.org/10.13140/RG.2.2.34414.61760), presented at the [e-Science Academy in Lund, on Oct 12-13 2016](essenceofescience.se/event/swedish-e-science-academy-2016-2/).
- See [slides from a recent presentation of SciPipe for use in a Bioinformatics setting](http://www.slideshare.net/SamuelLampa/scipipe-a-lightweight-workflow-library-inspired-by-flowbased-programming).
- The architecture of SciPipe is based on an [flow-based
  programming](https://en.wikipedia.org/wiki/Flow-based_programming) like
  pattern in pure Go presented in
  [this](http://blog.gopheracademy.com/composable-pipelines-pattern) and
  [this](https://blog.gopheracademy.com/advent-2015/composable-pipelines-improvements/)
  blog posts on Gopher Academy.

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
