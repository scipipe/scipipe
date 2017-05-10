# Basic concepts

In SciLuigi, we are discussing a few concepts all the time, so to make sure we
are on the same page, we will below go through the basic ones briefly.

## Processes

The probably most basic concept in SciPipe is the process.  A process is an
asynchronously running component that is typically defined as a static,
"long-running" part of the workflow, and the number of processes thus is
typically fixed for a workflow during its execution.

One can create customized types of processes, but for most basic workflows, the
[`scipipe.SciProcess`](https://godoc.org/github.com/scipipe/scipipe#SciProcess)
will be used, which is specialized for executing commandline applications. New
`SciProcess`-es are typically created using the `scipipe.NewFromShell(procName,
shellPattern)` command.

* See [GoDoc for SciProcess](https://godoc.org/github.com/scipipe/scipipe#SciProcess)

## Tasks

The "long-running" processes mentioned above, will receive input files on its
in-ports, and for each complete set of input files it receives, it will create
a new **task**. Specifically, `scipipe.SciProcess` will create
[`scipipe.SciTask`](https://godoc.org/github.com/scipipe/scipipe#SciTask) objects, and populate it with all data needed for one
particular shell command execution.  `SciTask` objects are executed via their
[`Execute()`](https://godoc.org/github.com/scipipe/scipipe#SciTask.Execute)
method, or `CustomExecute()`, if custom Go code is supposed to be
executed instead of a shell command.

The distinction between processes and tasks is important to understand, for
example when doing more advanced configuration of file naming strategies, since
the custom anonymous functions used to format paths are taking a `SciTask` as
input, even though these functions are saved on the process object.

To understand the difference between processes and tasks, it is helpful to
remember that processes are long-running, and typically fixed during the course
of a workflow, while tasks are transient objects, created temporarily as a
container for all data and code needed for each execution of a concrete shell
command.

* See [GoDoc for SciTask](https://godoc.org/github.com/scipipe/scipipe#SciTask)

## Ports

Central to the way data dependencies are defined in SciPipe, is ports. Ports
are fields on processes, which are connected to other ports via channels (see
separate section on this page).

In SciPipe, each port must have a unique name within its process (there can't
be an in-port and out-port named the same), and this name will be used in shell
command patterns, when connecting dependencies / dataflow networks, and when
configuring file naming strategies.

In `SciProcess` objects, in-ports are stored in a string->Port map field named
`In` (so they are accessed with: `myProcess.In["myport"]`), and out-ports
similarly in a string->Port map field named `Out`. Both are of type [`FilePort`](https://godoc.org/github.com/scipipe/scipipe#FilePort).

Some pre-made components might have ports bound to custom field names though,
such as `myFastaReader.InFastaFile`, or `myZipComponent.OutZipFile`.

Port objects have some methods bound to them, most importantly the `Connect()`
method, which takes another port, and connects to it, by stitching a channel
between the ports.

On `SciProcess` objects, there is also a third port type, `ParamPorts`, which
is used when it is needed to send a stream of parameter values (in string
format) to be supplied to as arguments to shell commands.

* See [GoDoc for the Port interface](https://godoc.org/github.com/scipipe/scipipe#Port)
* See [GoDoc for the FilePort struct type](https://godoc.org/github.com/scipipe/scipipe#FilePort)
* See [GoDoc for the ParamPort struct type](https://godoc.org/github.com/scipipe/scipipe#ParamPort)

## Channels

Ports in SciPipe are connected via channels. Channels are [plain Go channels](https://tour.golang.org/concurrency/2)
and nothing more. Most of the time, one will not need to deal with the channels
directly though, since the port objects (see separate section for ports) have
all the logic to connect to other ports via channels, but it can be good to
know that they are there, in case you need to do something more advanced.

## Pipeline runner

The [`PipelineRunner`](https://godoc.org/github.com/scipipe/scipipe#PipelineRunner)
is a special object in SciPipe, that just takes care of running a pipeline of
components. 

There is not much to say about the pipeline runner other than that it is
created with `scipipe.NewPipelineRunner()`, that all processes need to be added
to it in the right order (the last process last) with
`runner.AddProcesses(processes...)` and that it should be run with
`runner.Run()`. But this is already covered in the other examples and
tutorials.

* See [GoDoc for PipelineRunner](https://godoc.org/github.com/scipipe/scipipe#PipelineRunner)

## Shell command pattern

The `SciProcess` has the speciality that it can be configured using a special
shell command pattern, supplied to the [`NewFromShell()`](https://godoc.org/github.com/scipipe/scipipe#NewFromShell)
factory function. It is already explained in the section "writing workflows",
but in brief, it is a normal shell command, with placeholders for in-ports,
out-ports and parameter ports, on the form `{i:inportname}`, `{o:outportname}`,
and `{p:paramportname}`, respectively.

* See [GoDoc for NewFromShell()](https://godoc.org/github.com/scipipe/scipipe#NewFromShell)
