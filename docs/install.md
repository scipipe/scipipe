# Installing SciPipe

Installing SciPipe means first installing the Go programming langauge, and then
using Go's `go install` command to install the SciPipe library. After this, you will
be able to use Go's `go run` command to run SciPipe workflows.

## Install Go

Install Go by following the instructions [on this page](https://golang.org/doc/install#install), for your operating system.

To make sure that everything is installed, run the `go` command in your
terminal, and make sure that it outputs something.

To be specific, you can try exetuging `go version`, which sould output something
like the below:

```bash
$ go version
go version go1.17 linux/amd64
```
  
## Install SciPipe

There are two main ways of installing SciPipe, one which is super-easy, and one
which is recommended if you want to make sure that your workflow will never
break because of API changes in SciPipe, and that you always have a copy of the
SciPipe source code available.

### Easy: Using go install

The easiest way to intsall SciPipe is by using the `go install` tool in the Go
tool chain. To install scipipe with `go install`, run the following command in
your terminal:

```bash
go install github.com/scipipe/scipipe/...@latest
```

**N.B:** Don't miss the `...`, as otherwise the `scipipe` helper tool will not be installed.

## Initialize a new workflow file

Now, you should be able to write code like in the example below, in files
ending with `.go`.

The easiest way to get started is to let the scipipe tool generate a starting point for you:

```bash
scipipe new myfirstworkflow.go
```

... which you can then edit to your liking.

## Create a Go module

Before you can run the workflow, you need to also create a [go module](https://golang.org/ref/mod#introduction).

To do this, you can run the following command, in the directory where you
created your first workflow:

```bash
go mod init <package-name>
```

For `<module-name>`, you have to replace it with a name of the package.
For a simple script, you can name whatever you want, but if you are thinking
about publishing it online, e.g. on GitHub, you typically want to name it like
the URL of the corresponding GitHub repo, e.g.
`github.com/<your-username>/<your-repository>`.

By doing this, two files will be created:

```
go.mod
go.sum
```

Make sure to add them to your git repository, with:

```
git add go.mod go.sum
git commit -m "Add Go module files"
```

Now, to make sure that scipipe is included as a dependency in the go.mod file,
run the `go mod tidy` command:

```bash
go mod tidy
```

The `go.mod` file should now look something like:

```
module mylittlemodule

go 1.17

require github.com/scipipe/scipipe v0.10.2
```

... and the go.sum file might look something like:

```
github.com/scipipe/scipipe v0.10.2 h1:crXD1gGh/LuBfWfT4CdXcRFtPjem5weyXN03BDfVOuU=
github.com/scipipe/scipipe v0.10.2/go.mod h1:Nwof+Uimtam7GTpkU6cAf/EOnqvxcOVFytjnYU5I3vY=
```

### Optional extra step: Use a copy of SciPipe's source code in your own code

In order to make sure that your workflow will never break because of API
changes in SciPipe, and that you always have a copy of the SciPipe source
code available, we recommend to always include a copy of the SciPipe source
code in your workflow's source code repository. The SciPipe source code is
only around 1500 lines of code, with no external dependencies except Go and
Bash, so this should not increase the size of your repository too much.

A simple way to do this, is to use Go's `vendor` tool, which stores a local
copy of the source code of packages used, inside the local directory, in
a sub-directory called "vendor".

To do this, execute the following command:

```
go mod vendor
```

Then, to make sure the code is included in your git history, make sure to add it
to git:

```bash
git add vendor
git commit -m "Add vendored version of SciPipe"
```

## Run your workflow

Now youa re ready to run the workflow. To run a `.go` file, just use `go run <script-file>`, e.g:
  
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

There are also popular Go-plugins for [Sublime text](https://www.sublimetext.com), [Atom](https://atom.io/) and [IntelliJ IDEA](https://www.jetbrains.com/idea/), and an upcoming Go IDE from JetBrains,
called
