
<img src="images/fbp_factory.png" style="width: 30%; float: right; margin: 1em;">
<h1 style="margin-bottom: 0;"><img src="images/scipipe_logo_bluegrey_horiz.svg" style="width: 240px; margin-left: -10px; margin-bottom: 0;" alt="SciPipe"></h1>

<big>Robust, flexible and resource-efficient pipelines using Go and the commandline</big>

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

<strong><small>Project links: [GitHub repo](http://github.com/scipipe/scipipe) | [Issue Tracker](https://github.com/scipipe/scipipe/issues) | [Chat](https://gitter.im/scipipe/scipipe)</small></strong>

[![Build Status](https://img.shields.io/circleci/project/github/scipipe/scipipe.svg)](https://circleci.com/gh/scipipe/scipipe)
[![Test Coverage](https://img.shields.io/codecov/c/github/scipipe/scipipe.svg)](https://codecov.io/gh/scipipe/scipipe)
[![Codebeat Grade](https://codebeat.co/badges/96e93624-2ac8-42c9-9e94-2d6e5325d8ff)](https://codebeat.co/projects/github-com-scipipe-scipipe-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/scipipe/scipipe)](https://goreportcard.com/report/github.com/scipipe/scipipe)
[![GoDoc](https://godoc.org/github.com/scipipe/scipipe?status.svg)](https://godoc.org/github.com/scipipe/scipipe)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/scipipe/scipipe)
[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.1157941.svg)](https://doi.org/10.5281/zenodo.1157941)

<div style="width:37%; float: right; font-size: 0.8em; padding: 1em; margin: 0 0 1em 1em;">
<h4>Project updates</h4>
<strong>NEW - The SciPipe paper published open access in GigaScience:</strong> [SciPipe: A workflow library for agile development of complex and dynamic bioinformatics pipelines](https://doi.org/10.1093/gigascience/giz044)
<strong>Scientific study using SciPipe:</strong> [Predicting off-target binding profiles with confidence using Conformal Prediction](https://doi.org/10.3389/fphar.2018.01256)
<strong>Slides:</strong> [Presentation on SciPipe and more at Go Stockholm Conference](https://pharmb.io/blog/saml-gostockholm2018/)
<strong>Blog post:</strong> [Provenance reports in Scientific Workflows](http://bionics.it/posts/provenance-reports-in-scientific-workflows) - going into details about how SciPipe is addressing provenance
<strong>Blog post:</strong> [First production workflow run with SciPipe](http://bionics.it/posts/first-production-workflow-run-with-scipipe)
</div>

## Introduction

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

## Known limitations

- There are still a number of missing good-to-have features for workflow
  design. See the [issue tracker](https://github.com/scipipe/scipipe/issues)
  for details.
- There is not (yet) support for the [Common Workflow Language](http://common-workflow-language.github.io).

## Installing

1. [Download](https://golang.org/dl/) and [install](https://golang.org/doc/install) Go
2. Run the following command, to install the scipipe Go library (don't miss the trailing dots!):

```bash
go get github.com/scipipe/scipipe/...
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

## Running the example

Let's put the code in a file named `scipipe_helloworld.go` and run it:

```bash
$ go run hello_world.go
AUDIT   2018/06/15 19:04:22 | workflow:hello_world             | Starting workflow (Writing log to log/scipipe-20180615-190422-hello_world.log)
AUDIT   2018/06/15 19:04:22 | hello                            | Executing: echo 'Hello ' > hello.out.txt.tmp/hello.out.txt
AUDIT   2018/06/15 19:04:22 | hello                            | Finished:  echo 'Hello ' > hello.out.txt.tmp/hello.out.txt
AUDIT   2018/06/15 19:04:22 | world                            | Executing: echo $(cat hello.out.txt) World > hello.out.txt.world.out.txt.tmp/hello.out.txt.world.out.txt
AUDIT   2018/06/15 19:04:22 | world                            | Finished:  echo $(cat hello.out.txt) World > hello.out.txt.world.out.txt.tmp/hello.out.txt.world.out.txt
AUDIT   2018/06/15 19:04:22 | workflow:hello_world             | Finished workflow (Log written to log/scipipe-20180615-190422-hello_world.log)
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

Now, these were a little long and cumbersome filename, weren't they? SciPipe
gives you very good control over how to name your files, if you don't want to
rely on the automatic file naming. For example, we could set the first filename
statically, and then use the first name as a basis for the file name for the
second process, like so:

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

For more information about how to write workflows using SciPipe, use the menu
to the left, to browse the various topics!
