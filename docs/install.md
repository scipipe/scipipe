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

## Some tips about editors

In order to be productive with SciPipe, you will also need a Go editor or IDE
with support for auto-completion, sometimes also called "intellisense".

We can warmly recommend to use one of these editors, sorted by level of endorsement:

1. [Visual Studio Code](http://code.visualstudio.com) with the [Go plugin](https://github.com/Microsoft/vscode-go) - If you want a very powerful almost IDE-like editor
2. Fatih's awesome [vim-go](https://github.com/fatih/vim-go) plugin - if you are a Vim power-user
3. [LiteIDE](https://github.com/visualfc/liteide) - if you want a really simple, standalone Go-editor

There are also popular Go-plugins for [Sublime text](https://www.sublimetext.com),
[Atom](https://atom.io/) and [IntelliJ IDEA](https://www.jetbrains.com/idea/),
and an upcoming Go IDE from JetBrains, called
[Gogland](https://www.jetbrains.com/go/), that might be worth checking out,
depending on your preferences.
