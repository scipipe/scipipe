# Contributing to SciPipe

## Working with forked repository

Unlike other languages like C++/Python, Go requires modules being referenced,
to be placed in a specific location for the import to work. Forking via GitHub
provides a pull request workflow that is well documented but does not work well
with Go import. Documented here is one approach utilizing go modules.

First, fork and clone SciPipe the usual way via GitHub to a local file system,
lets call it `<cloned-scipipe-dir>`

To work on and test the changes made to scipipe in the cloned location, create
a directory for your scipipe workflows, lets call it `<workflow-dev-dir>`. In
that directory, you can create a simple `main.go` file with a `package main` so
that you can run your code that exercises the changes you are making at
`<cloned-scipipe-dir>`.

Note that if you were to run your code in `main.go` as-is, it would pull down
the code from the repository and cache them and you will not actually be able
to test any local changes in scipipe. To do so, you need to create a file call
`go.mod`.

In `<workflow-dev-dir>` run the following:

```
go mod init <some package name>
```

Note: The actual name of the package is not critical.

Next, define the modules required by your main.go:

```
go mod edit -require=github.com/scipipe/scipipe@v0.9.10
```

Note: A the time of this writing, v0.9.10 is the latest published version, this
will change over time and you need to adapt.

Next, replace any reference to the previous URL with reference to the
`<cloned-scipipe-dir>` location:

```
go mod edit -replace=github.com/scipipe/scipipe@v0.9.10=<cloned-scipipe-dir>
```

Now when you do:

```
go run main.go
```

... it will not be pulling the code from GitHub but references the code you
have cloned locally.

Do your development work and push to your forked repository and do a pull
request for the author to review and optionally merge the contribution.
