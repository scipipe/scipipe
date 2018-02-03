# Basic concepts

In SciPipe, we are discussing a few concepts all the time, so to make sure we
are on the same page, we will below go through the basic ones briefly.

## Processes

The probably most basic concept in SciPipe is the process.  A process is an
asynchronously running component that is typically defined as a static,
"long-running" part of the workflow, and the number of processes thus is
typically fixed for a workflow during its execution.

One can create customized types of processes, but for most basic workflows, the
[`scipipe.Process`](https://godoc.org/github.com/scipipe/scipipe#Process)
will be used, which is specialized for executing commandline applications. New
`Process`-es are typically created using the `scipipe.NewProc(procName,
shellPattern)` command.

* See [GoDoc for Process](https://godoc.org/github.com/scipipe/scipipe#Process)

## Tasks

The "long-running" processes mentioned above, will receive input files on its
in-ports, and for each complete set of input files it receives, it will create
a new **task**. Specifically, `scipipe.Process` will create
[`scipipe.Task`](https://godoc.org/github.com/scipipe/scipipe#Task) objects, and populate it with all data needed for one
particular shell command execution.  `Task` objects are executed via their
[`Execute()`](https://godoc.org/github.com/scipipe/scipipe#Task.Execute)
method, or `CustomExecute()`, if custom Go code is supposed to be
executed instead of a shell command.

The distinction between processes and tasks is important to understand, for
example when doing more advanced configuration of file naming strategies, since
the custom anonymous functions used to format paths are taking a `Task` as
input, even though these functions are saved on the process object.

To understand the difference between processes and tasks, it is helpful to
remember that processes are long-running, and typically fixed during the course
of a workflow, while tasks are transient objects, created temporarily as a
container for all data and code needed for each execution of a concrete shell
command.

* See [GoDoc for Task](https://godoc.org/github.com/scipipe/scipipe#Task)

## Ports

Central to the way data dependencies are defined in SciPipe, is ports. Ports
are fields on processes, which are connected to other ports via channels (see
separate section on this page).

In SciPipe, each port must have a unique name within its process (there can't
be an in-port and out-port named the same), and this name will be used in shell
command patterns, when connecting dependencies / dataflow networks, and when
configuring file naming strategies.

In `Process` objects, in-ports are are accessed with
`myProcess.In("my_port")`, and out-ports are similarly accessed with
`myProcess.Out("my_other_port")`. They are of type
[`InPort`](https://godoc.org/github.com/scipipe/scipipe#InPort) and
[`OutPort`](https://godoc.org/github.com/scipipe/scipipe#OutPort) respectively.

Some pre-made components might have ports bound to custom field names though,
such as `myFastaReader.InFastaFile`, or `myZipComponent.OutZipFile`.

Port objects have some methods bound to them, most importantly the `Connect()`
method, which takes another port, and connects to it, by stitching a channel
between the ports.

On `Process` objects, there is also a third port type, `ParamInPort` (and the
accompanying `ParamOutPort`), which is used when it is needed to send a
stream of parameter values (in string format) to be supplied to as arguments
to shell commands.

* See [GoDoc for the InPort struct type](https://godoc.org/github.com/scipipe/scipipe#InPort)
* See [GoDoc for the OutPort struct type](https://godoc.org/github.com/scipipe/scipipe#OutPort)
* See [GoDoc for the ParamInPort struct type](https://godoc.org/github.com/scipipe/scipipe#ParamInPort)
* See [GoDoc for the ParamOutPort struct type](https://godoc.org/github.com/scipipe/scipipe#ParamOutPort)

## Channels

Ports in SciPipe are connected via channels. Channels are [plain Go channels](https://tour.golang.org/concurrency/2)
and nothing more. Most of the time, one will not need to deal with the channels
directly though, since the port objects (see separate section for ports) have
all the logic to connect to other ports via channels, but it can be good to
know that they are there, in case you need to do something more advanced.

## Workflow

The [`Workflow`](https://godoc.org/github.com/scipipe/scipipe#Workflow)
is a special object in SciPipe, that just takes care of running a set of
components making up a workflow.

There is not much to say about the workflow component, other than that it is
created with `scipipe.NewWorkflow(workflowName, maxConcurrentTasks)`, that all processes need to be added
to it with `wf.AddProc(proc)` while the "last", or "driving" process needs to be specified with `wf.SetDriver(driverProcess)`, and that it should be run with
`wf.Run()`. But this is already covered in the other examples and
tutorials.

* See [GoDoc for Workflow](https://godoc.org/github.com/scipipe/scipipe#Workflow)

## Shell command pattern

The `Process` has the speciality that it can be configured using a special
shell command pattern, supplied to the [`NewProc()`](https://godoc.org/github.com/scipipe/scipipe#NewProc)
factory function. It is already explained in the section "writing workflows",
but in brief, it is a normal shell command, with placeholders for in-ports,
out-ports and parameter ports, on the form `{i:inportname}`, `{o:outportname}`,
and `{p:paramportname}`, respectively.

* See [GoDoc for NewProc()](https://godoc.org/github.com/scipipe/scipipe#NewProc)
