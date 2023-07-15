<h1 style="margin-bottom: 0;"><img src="docs/images/scipipe_logo_bluegrey_horiz_320px.png" alt="SciPipe"></h1>

<big>Robust, flexible and resource-efficient pipelines using Go and the commandline</big>

[![Build Status](https://img.shields.io/circleci/project/github/scipipe/scipipe.svg)](https://app.circleci.com/pipelines/github/scipipe)
[![Test Coverage](https://img.shields.io/codecov/c/github/scipipe/scipipe.svg)](https://codecov.io/gh/scipipe/scipipe)
[![Codebeat Grade](https://codebeat.co/badges/96e93624-2ac8-42c9-9e94-2d6e5325d8ff)](https://codebeat.co/projects/github-com-scipipe-scipipe-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/scipipe/scipipe)](https://goreportcard.com/report/github.com/scipipe/scipipe)
[![GoDoc](https://godoc.org/github.com/scipipe/scipipe?status.svg)](https://godoc.org/github.com/scipipe/scipipe)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/scipipe/scipipe)
[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.1157941.svg)](https://doi.org/10.5281/zenodo.1157941)

<strong>Project links: [Documentation & Main Website](http://scipipe.org) | [Issue Tracker](https://github.com/scipipe/scipipe/issues) | [Chat](https://gitter.im/scipipe/scipipe)</strong>

## Why SciPipe?

- **Intuitive:** SciPipe works by flowing data through a network of channels
  and processes
- **Flexible:** Wrapped command-line programs can be combined with processes in
  Go
- **Convenient:** Full control over how your files are named
- **Efficient:** Workflows are compiled to binary code that run fast
- **Parallel:** Pipeline paralellism between processes as well as task
  parallelism for multiple inputs, making efficient use of multiple CPU cores
- **Supports streaming:** Stream data between programs to avoid wasting disk space
- **Easy to debug:** Use available Go debugging tools or just `println()`
- **Portable:** Distribute workflows as Go code or as self-contained executable
  files


## Project updates

- <strong>Jan 2020: New screencast:</strong> <a href="https://www.youtube.com/watch?v=hi0Uqwddrtg" target="_blank">"Hello World" scientific workflow in SciPipe</a>
- <strong>May 2019: The SciPipe paper published open access in GigaScience:</strong> <a href="https://doi.org/10.1093/gigascience/giz044" target="_blank">SciPipe: A workflow library for agile development of complex and dynamic bioinformatics pipelines</a>
- <strong>Nov 2018: Scientific study using SciPipe:</strong> <a href="https://doi.org/10.3389/fphar.2018.01256" target="_blank">Predicting off-target binding profiles with confidence using Conformal Prediction</a>
- <strong>Slides:</strong> <a href="https://pharmb.io/blog/saml-gostockholm2018/">Presentation on SciPipe and more at Go Stockholm Conference</a>
- <strong>Blog post:</strong> <a href="http://bionics.it/posts/provenance-reports-in-scientific-workflows">Provenance reports in Scientific Workflows</a> - going into details about how SciPipe is addressing provenance.
- <strong>Blog post:</strong> <a href="http://bionics.it/posts/first-production-workflow-run-with-scipipe">First production workflow run with SciPipe</a

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

**Project status:** SciPipe pretty stable now, and only very minor API changes
might still occur. We have successfully used SciPipe in a handful of both real
and experimental projects, and it has had occasional use outside the research
group as well.

## Known limitations

- There are still a number of missing good-to-have features for workflow
  design. See the [issue tracker](https://github.com/scipipe/scipipe/issues)
  for details.
- There is not (yet) support for the [Common Workflow Language](http://common-workflow-language.github.io).

## Installing

For full installation instructions, see the [intallation page](https://scipipe.org/install/).
For quick getting started steps, you can do:

1. [Download](https://golang.org/dl/) and [install](https://golang.org/doc/install) Go
2. Run the following command, to install the scipipe Go library (don't miss the
   trailing dots!), and create a Go module for your script:

```bash
go install github.com/scipipe/scipipe/...@latest
go mod init myfirstworkflow-module
```

## Hello World example

Let's look at an example workflow to get a feel for what writing workflows in
SciPipe looks like:

```go
package main

import (
    // Import SciPipe, aliased to sp
    sp "github.com/scipipe/scipipe"
)

func main() {
    // Init workflow and max concurrent tasks
    wf := sp.NewWorkflow("hello_world", 4)

    // Initialize processes, and file extensions
    hello := wf.NewProc("hello", "echo 'Hello ' > {o:out|.txt}")
    world := wf.NewProc("world", "echo $(cat {i:in}) World > {o:out|.txt}")

    // Define data flow
    world.In("in").From(hello.Out("out"))

    // Run workflow
    wf.Run()
}
```

To create a file with a similar simple example, you can run:

```
scipipe new hello_world.go
```

## Running the example

Let's put the code in a file named `hello_world.go` and run it.

First you need to make sure that the dependencies (SciPipe in this case) is
installed in your local Go module. This you can do with:

```bash
go mod tidy
```

Then you can go ahead and run the workflow:

```bash
$ go run hello_world.go
AUDIT   2018/07/17 21:42:26 | workflow:hello_world             | Starting workflow (Writing log to log/scipipe-20180717-214226-hello_world.log)
AUDIT   2018/07/17 21:42:26 | hello                            | Executing: echo 'Hello ' > hello.out.txt
AUDIT   2018/07/17 21:42:26 | hello                            | Finished: echo 'Hello ' > hello.out.txt
AUDIT   2018/07/17 21:42:26 | world                            | Executing: echo $(cat ../hello.out.txt) World > hello.out.txt.world.out.txt
AUDIT   2018/07/17 21:42:26 | world                            | Finished: echo $(cat ../hello.out.txt) World > hello.out.txt.world.out.txt
AUDIT   2018/07/17 21:42:26 | workflow:hello_world             | Finished workflow (Log written to log/scipipe-20180717-214226-hello_world.log)
```

Let's check what file SciPipe has generated:

```
$ ls -1 hello*
hello.out.txt
hello.out.txt.audit.json
hello.out.txt.world.out.txt
hello.out.txt.world.out.txt.audit.json
```

As you can see, it has created a file `hello.out.txt`, and `hello.out.world.out.txt`, and
an accompanying `.audit.json` for each of these files.

Now, let's check the output of the final resulting file:

```bash
$ cat hello.out.txt.world.out.txt
Hello World
```

Now we can rejoice that it contains the text "Hello World", exactly as a proper
Hello World example should :)

Now, these were a little long and cumbersome filenames, weren't they? SciPipe
gives you very good control over how to name your files, if you don't want to
rely on the automatic file naming. For example, we could set the first filename
to a static one, and then use the first name as a basis for the file name for
the second process, like so:

```go
package main

import (
    // Import the SciPipe package, aliased to 'sp'
    sp "github.com/scipipe/scipipe"
)

func main() {
    // Init workflow with a name, and max concurrent tasks
    wf := sp.NewWorkflow("hello_world", 4)

    // Initialize processes and set output file paths
    hello := wf.NewProc("hello", "echo 'Hello ' > {o:out}")
    hello.SetOut("out", "hello.txt")

    world := wf.NewProc("world", "echo $(cat {i:in}) World >> {o:out}")
    world.SetOut("out", "{i:in|%.txt}_world.txt")

    // Connect network
    world.In("in").From(hello.Out("out"))

    // Run workflow
    wf.Run()
}
```

In the `{i:in...` part, we are re-using the file path from the file received on
the in-port named 'in', and then running a Bash-style trim-from-end command on
it to remove the `.txt` extension.

Now, if we run this, the file names get a little cleaner:

```bash
$ ls -1 hello*
hello.txt
hello.txt.audit.json
hello_world.go
hello_world.txt
hello_world.txt.audit.json
```

## The audit logs

Finally, we could have a look at one of those audit file created:

```bash
$ cat hello_world.txt.audit.json
{
    "ID": "99i5vxhtd41pmaewc8pr",
    "ProcessName": "world",
    "Command": "echo $(cat hello.txt) World \u003e\u003e hello_world.txt.tmp/hello_world.txt",
    "Params": {},
    "Tags": {},
    "StartTime": "2018-06-15T19:10:37.955602979+02:00",
    "FinishTime": "2018-06-15T19:10:37.959410102+02:00",
    "ExecTimeNS": 3000000,
    "Upstream": {
        "hello.txt": {
            "ID": "w4oeiii9h5j7sckq7aqq",
            "ProcessName": "hello",
            "Command": "echo 'Hello ' \u003e hello.txt.tmp/hello.txt",
            "Params": {},
            "Tags": {},
            "StartTime": "2018-06-15T19:10:37.950032676+02:00",
            "FinishTime": "2018-06-15T19:10:37.95468214+02:00",
            "ExecTimeNS": 4000000,
            "Upstream": {}
        }
    }
```

Each such audit-file contains a hierarchic JSON-representation of the full
workflow path that was executed in order to produce this file. On the first
level is the command that directly produced the corresponding file, and then,
indexed by their filenames, under "Upstream", there is a similar chunk
describing how all of its input files were generated. This process will be
repeated in a recursive way for large workflows, so that, for each file
generated by the workflow, there is always a full, hierarchic, history of all
the commands run - with their associated metadata - to produce that file.

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

## Citing SciPipe

If you use SciPipe in academic or scholarly work, please cite the following paper as source:

Lampa S, Dahlö M, Alvarsson J, Spjuth O. SciPipe: A workflow library for agile development of complex and dynamic bioinformatics pipelines 
_Gigascience_. 8, 5 (2019). DOI: [10.1093/gigascience/giz044](https://dx.doi.org/10.1093/gigascience/giz044)

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
