# Installing SciPipe

Installing SciPipe means first installing the Go programming langauge, and then
using Go's `go get` command to install the SciPipe library. After this, you will
be able to use Go's `go run` command to run SciPipe workflows.

## Install Go

Install Go by following the instructions [on this page](https://golang.org/doc/install#install),
for your operating system.
  
## Install SciPipe

There are two main ways of installing SciPipe, one which is maximally easy, and one which
is recommended if you want to make sure that your workflow will never break because of
API changes in SciPipe, and that you always have a copy of the SciPipe source code available.

### Easiest: Using go get

The easiest way to intsall SciPipe is by using the `go get` tool in the Go
tool chain. To install scipipe with `go get`, run the following command in
your terminal:

```bash
go get github.com/scipipe/scipipe/...
```

**N.B:** Don't miss the `...`, as otherwise the `scipipe` helper tool will not be installed.

### For maximum future proofing: Use a copy of SciPipe's source code in your own code

In order to make sure that your workflow will never break because of API
changes in SciPipe, and that you always have a copy of the SciPipe source
code available, we recommend to always include a copy of the SciPipe source
code in your workflow's source code repository. The SciPipe source code is
only around 1500 lines of code, with no external dependencies except Go and
Bash, so this should not increase the size of your repository too much.

A simple way to do this, is to clone a copy of the SciPipe source code into a
folder structure that looks like this, under your main workflow code folder
(where you store your own `.go` files):

```
vendor/src/github.com/scipipe/scipipe
```

To create and clone the scipipe repo to this folder, you can use these
commands:

```bash
mkdir -p vendor/src/github.com/scipipe
cd vendor/src/github.com/scipipe
git clone https://github.com/scipipe/scipipe.git
```

## Initialize a new workflow file
Now, you should be able to write code like in the example below, in files
ending with `.go`.

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
2. The [vim-go](https://github.com/fatih/vim-go) plugin by [Fatih](https://twitter.com/fatih) - if you are a Vim power-user, or need a terminal-only complement to VSCode.
3. JetBrain's [GoLand IDE](https://www.jetbrains.com/go/), if you are ready to pay for maximum code intelligence in a professional IDE.
4. [LiteIDE](https://github.com/visualfc/liteide) - if you want a simple, robust and fast standalone Go-editor.

There are also popular Go-plugins for [Sublime text](https://www.sublimetext.com),
[Atom](https://atom.io/) and [IntelliJ IDEA](https://www.jetbrains.com/idea/),
and an upcoming Go IDE from JetBrains, called