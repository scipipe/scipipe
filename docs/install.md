# Install

## Install Go

First install Go by following instructions on [this page](https://golang.org/doc/install).
  - I typically install to a custom location (`~/go` for the go tools, and `~/code/go` for my own go-projects).
  - If you want to install (which means, untar the go tarball) to `~/go` just like me, you should put the following in your `~/.bashrc` file:
  
```bash
# Go stuff
export GOROOT=~/go
export GOPATH=~/code/go
export PATH=$GOROOT/bin:$PATH
export PATH=$GOPATH/bin:$PATH
```
  
## Install SciPipe

Install SciPipe by running the following shell command:
  
```bash
go get github.com/scipipe/scipipe/...
```

**N.B:** Don't miss the `...`, or you won't get the `scipipe` helper tool.
  
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
