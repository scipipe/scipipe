# Installing SciPipe

Installing SciPipe means first installing the Go programming langauge, and then
using Go's `go get` command to install the SciPipe library. After this, you will
be able to use Go's `go run` command to run SciPipe workflows.

## Install Go

Install Go by following the instructions [on this page](https://golang.org/doc/install#install),
for your operating system.
  
## Install SciPipe

Then install SciPipe by running the following shell command:
  
```bash
go get github.com/scipipe/scipipe/...
```

**N.B:** Don't miss the `...`, as otherwise the `scipipe` helper tool will not be installed.
  
## Initialize a new workflow file
  
Now, you should be able to write code like in the example below, in files ending with `.go`.

The easiest way to get started is to let the scipipe tool generate a starting point for you:

```bash
scipipe new myfirstworkflow.go
```

... which you can then edit to your liking.

## Run your workflow

To run a `.go` file, use `go run`:
  
```bash
go run myfirstworkflow.go
```
