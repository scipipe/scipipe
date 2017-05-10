# Basic concepts

## Defining workflows

Defining workflows in general (not unique to SciPipe) typically means:

1. Defining the processes (or "tasks" or "components") of the workflow.
2. Defining dependencies between the processes.
3. Running the workflow.

### Minimal boilerplate code

In Scipipe, everything is normal Go code, so SciPipe workflows are also
full, normal Go programs. These always need to do three things:

1. Define the `main` package (always te case for executable files)
2. Import the SciPipe library
3. Have a `main()` function, which contains the workflow code.

Thus, the minimal "boilerplate" code for any SciPipe workflow is a file
named with the `.go` extension, and with this content:

```go
package main

import (
    "github.com/scipipe/scipipe"
)

func main() {
    // All workflow code goes here
}
```

Then, this file will be runnable with the `go run` command, like:

```bash
go run myworkflow.go
```

### Aliasing scipipe for less typing

If you find it tedious to type `scipipe` over and over in your code you can alias
it to something shorter, like `sp`, by doing the import like this:

```go
import (
    sp "github.com/scipipe/scipipe"
)
```

## Defining processes

In SciPipe, processes are defined using the `NewFromShell()` command, by
providing a process name and a shell pattern, where file names are replaced
with place-holders with a port-name. Just like so:

```go
myProcess := scipipe.NewFromShell("myprocess", "echo hi > {o:outfile}")
```

Based on this shell command pattern, a process is created, with in- and
out-ports and where the shell command to be executed, will be calculated from
the provided pattern.

## Defining dependencies

The dependency definitions in SciPipe are done by "physically" connecting ports
(in-ports and out-ports) to each other via buffered channels, on which data
objects will travel between processes. The connection is done in practice with
a `Connect()` method, available on each port object, which takes another port
object as input, in order to connect the two ports.

Ports created when using the shell pattern, are stored in the fields `In` and
`Out` on each process, under their own name, since the fields are maps.  So, an
out-port named "outfile" will be accessed from `myProcess` with:
`myProcess.Out["outfile"]`, and and in-port named "inport" will be accessed from
`myOhterProcess` with: `myOtherProcess.In["infile"]`.

Connecting `myOtherProcess.In["infile"]` with `myProcess.Out["outfile"]` is done
simply with:

```go
myOtherProcess.In["infile"].Connect(myProcess.Out["outfile"])
```

... or

```go
myProcess.Out["outfile"].Connect(myOtherProcess.In["infile"])
```

... since the operation is symmetric.

There is also an alternative syntax using the `scipipe.Connect()` method, which
takes two port-objects and connects them:

```go
scipipe.Connect(myOtherProcess.In["infile"], myProcess.Out["outfile"])
```

... or

```go
scipipe.Connect(myProcess.Out["outfile"], myOtherProcess.In["infile"])
```

## Running workflows

In order to run SciPipe workflows, you have to add all processes to a pipeline runner,
and then execute the `Run()` method on the pipeline runner, like so:

```go
runner := scipipe.NewPipelineRunner()
runner.AddProcesses(myProcess, myOtherProcess)
runner.Run()
```

*Important:* Note that the order processes are added to the pipeline runner is
important!  The last process has to be added last, in order for the pipeline to
function properly.

# Next steps

Please see the [Hello World tutorial](/tutorials/helloworld/), for a concrete
example of using the concepts above.
