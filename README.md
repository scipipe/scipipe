# SciPipe

[![Build Status](https://img.shields.io/circleci/project/github/scipipe/scipipe.svg)](https://circleci.com/gh/scipipe/scipipe)
[![Test Coverage](https://img.shields.io/codecov/c/github/scipipe/scipipe.svg)](https://codecov.io/gh/scipipe/scipipe)
[![Codebeat Grade](https://codebeat.co/badges/96e93624-2ac8-42c9-9e94-2d6e5325d8ff)](https://codebeat.co/projects/github-com-scipipe-scipipe-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/scipipe/scipipe)](https://goreportcard.com/report/github.com/scipipe/scipipe)
[![GoDoc](https://godoc.org/github.com/scipipe/scipipe?status.svg)](https://godoc.org/github.com/scipipe/scipipe)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/scipipe/scipipe)
[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.1157941.svg)](https://doi.org/10.5281/zenodo.1157941)

<strong>Project links: [Documentation & Main Website](http://scipipe.org) | [Issue Tracker](https://github.com/scipipe/scipipe/issues) | [Mailing List](https://groups.google.com/forum/#!forum/scipipe)</strong>

## Project updates

- <strong>NEW blog post:</strong> [Provenance reports in Scientific Workflows](http://bionics.it/posts/provenance-reports-in-scientific-workflows) - going into details about how SciPipe is addressing provenance
- <strong>NEW blog post:</strong> [First production workflow run with SciPipe](http://bionics.it/posts/first-production-workflow-run-with-scipipe)
- <strong>NEW video:</strong> [Watch a screencast on how to write a Hello World workflow in SciPipe [15:28]](https://www.youtube.com/watch?v=kWqkGwDU-Hc)

## Introduction

<img src="docs/images/fbp_factory.png" align="right">

SciPipe is a library for writing [Scientific
Workflows](https://en.wikipedia.org/wiki/Scientific_workflow_system), sometimes
also called "pipelines", in the [Go programming language](http://golang.org).

When you need to run many commandline programs that depend on each other in
complex ways, SciPipe helps by making the process of running these programs
flexible, robust and reproducible. SciPipe also lets you restart an interrupted
run without over-writing already produced output and produces an audit report
of what was run, among many other things.

SciPipe is built on the proven principles of [Flow-Based Programming](https://en.wikipedia.org/wiki/Flow-based_programming)
(FBP) to achieve maximum flexibility, productivity and agility when designing
workflows.  Compared to plain dataflow, FBP provides the benefits that
processes are fully self-contained, so that a library of re-usable components
can be created, and plugged into new workflows ad-hoc.

Similar to other FBP systems, SciPipe workflows can be likened to a network of
assembly lines in a factory, where items (files) are flowing through a network
of conveyor belts, stopping at different independently running stations
(processes) for processing, as depicted in the picture above.

SciPipe was initially created for problems in bioinformatics and
cheminformatics, but works equally well for any problem involving pipelines of
commandline applications.

**Project status:** SciPipe is still alpha software and minor breaking API
changes still happens as we try to streamline the process of writing workflows.
Please follow the commit history closely for any API updates if you have code
already written in SciPipe (Let us know if you need any help in migrating code
to the latest API).

## Benefits

Some key benefits of SciPipe, that are not always found in similar systems:

- **Intuitive behaviour:** SciPipe operates by flowing data (files) through a
  network of channels and processes, not unlike the conveyor belts and stations
  in a factory.
- **Flexible:** Processes that wrap command-line programs or scripts, can be
  combined with processes coded directly in Golang.
- **Custom file naming:** SciPipe gives you full control over how files are
  named, making it easy to find your way among the output files of your
  workflow.
- **Portable:** Workflows can be distributed either as Go code to be run with
  `go run`, or as stand-alone executable files that run on almost any UNIX-like
  operating system.
- **Easy to debug:** As everything in SciPipe is just Go code, you can use some
  of the available debugging tools, or just `println()` statements, to debug
  your workflow. 
- **Supports streaming:** Can stream outputs via UNIX FIFO files, to avoid temporary storage.
- **Efficient and Parallel:** Workflows are compiled into statically compiled
  code that runs fast. SciPipe also leverages pipeline parallelism between
  processes as well as task parallelism when there are multiple inputs to a
  process, making efficient use of multiple CPU cores.

## Known limitations

- There are still a number of missing good-to-have features for workflow
  design. See the [issue tracker](https://github.com/scipipe/scipipe/issues)
  for details.
- There is not (yet) support for the [Common Workflow Language](http://common-workflow-language.github.io).

## Hello World example

Let's look at an example workflow to get a feel for what writing workflows in
SciPipe looks like:

```go
package main

import (
    // Import SciPipe into the main namespace (generally frowned upon but could
    // be argued to be reasonable for short-lived workflow scripts like this)
    . "github.com/scipipe/scipipe"
)

func main() {
    // Init workflow
    wf := NewWorkflow("hello_world")

    // Initialize processes and set output file paths
    hello := wf.NewProc("hello", "echo 'Hello ' > {o:out}")
    hello.SetPathStatic("out", "hello.txt")

    world := wf.NewProc("world", "echo $(cat {i:in}) World >> {o:out}")
    world.SetPathReplace("in", "out", ".txt", "_world.txt")

    // Connect network
    world.In("in").Connect(hello.Out("out"))

    // Run workflow
    wf.Run()
}
```

## Running the example

Let's put the code in a file named `scipipe_helloworld.go` and run it:

```bash
$ go run scipipe_helloworld.go 
AUDIT   2017/05/04 17:05:15 Task:hello        Executing command: echo 'Hello ' > hello.txt.tmp
AUDIT   2017/05/04 17:05:15 Task:world        Executing command: echo $(cat hello.txt) World >> hello_world.txt.tmp
```

Let's check what file SciPipe has generated:

```
$ ls -1tr hello*
hello.txt.audit.json
hello.txt
hello_world.txt
hello_world.txt.audit.json
```

As you can see, it has created a file `hello.txt`, and `hello_world.txt`, and
an accompanying `.audit.json` for each of these files.

Now, let's check the output of the final resulting file:

```bash
$ cat hello_world.txt
Hello World
```

Now we can rejoice that it contains the text "Hello World", exactly as a proper
Hello World example should :)

You can find many more examples in the [examples folder](https://github.com/scipipe/scipipe/tree/master/examples) in the GitHub repo.

For more information about how to write workflows using SciPipe, and much more,
see [SciPipe website (scipipe.org)](http://scipipe.org)!

## More material on SciPipe

- See [a poster on SciPipe](http://dx.doi.org/10.13140/RG.2.2.34414.61760), presented at the [e-Science Academy in Lund, on Oct 12-13 2016](essenceofescience.se/event/swedish-e-science-academy-2016-2/).
- See [slides from a recent presentation of SciPipe for use in a Bioinformatics setting](http://www.slideshare.net/SamuelLampa/scipipe-a-lightweight-workflow-library-inspired-by-flowbased-programming).
- The architecture of SciPipe is based on an [flow-based programming](https://en.wikipedia.org/wiki/Flow-based_programming) like
  pattern in pure Go presented in
  [this](http://blog.gopheracademy.com/composable-pipelines-pattern) and
  [this](https://blog.gopheracademy.com/advent-2015/composable-pipelines-improvements/)
  blog posts on Gopher Academy.

## Acknowledgements

- SciPipe is very heavily dependent on the proven principles form [Flow-Based
  Programming (FBP)](http://www.jpaulmorrison.com/fbp), as invented by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
  From Flow-based programming, SciPipe uses the ideas of separate network
  (workflow dependency graph) definition, named in- and out-ports,
  sub-networks/sub-workflows and bounded buffers (already available in Go's
  channels) to make writing workflows as easy as possible.
- This library is has been much influenced/inspired also by the
  [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- Thanks to [Egon Elbre](http://twitter.com/egonelbre) for helpful input on the
  design of the internals of the pipeline, and processes, which greatly
  simplified the implementation.
- This work is financed by faculty grants and other financing for the [Pharmaceutical Bioinformatics group](http://pharmb.io) of [Dept. of
  Pharmaceutical Biosciences](http://www.farmbio.uu.se) at [Uppsala University](http://www.uu.se), and by [Swedish Research Council](http://vr.se)
  through the Swedish [National Bioinformatics Infrastructure Sweden](http://nbis.se).
- Supervisor for the project is [Ola Spjuth](http://www.farmbio.uu.se/research/researchgroups/pb/olaspjuth).

## Related tools

Find below a few tools that are more or less similar to SciPipe that are worth worth checking out before
deciding on what tool fits you best (in approximate order of similarity to SciPipe):

- [NextFlow](http://nextflow.io)
- [Luigi](https://github.com/spotify/luigi)/[SciLuigi](https://github.com/samuell/sciluigi)
- [BPipe](https://code.google.com/p/bpipe/)
- [SnakeMake](https://bitbucket.org/johanneskoester/snakemake)
- [Cuneiform](https://github.com/joergen7/cuneiform)
