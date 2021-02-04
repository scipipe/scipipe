# Contributing to SciPipe

## Working with forked repository
Unlike other languages like C++/Python, golang requiries modules being referenced
to be place in specific location for the import to work. Forking via github
provides PR workflow that are well documented but does not work well with
golang import. Documented here is one approach utilizing the go.mod approach.

Fork and clone scipipe the usual way via github to a local file system, lets
call it <cloned-scipipe-dir>

To work on and test the changes made to scipipe in the cloned location, create
a directory, lets call it <scipipe-dev-dir>, in that directory, you can
create a simple main.go with a package main so that you can run your code that
exercises the changes you are making at <cloned-scipipe-dir>

Note that if you were to run your code in main.go as-is, it will pull down the code from the repository and cached them and you will not actually be able to test your code. To do so, you need to create a file call go.mod

In <scipipe-dev-dir> run the following

```
go mod init <some package name>
```
Note: The actual name of the package is not critical

Next, define the modules required by your main.go
```
go mod edit -require=github.com/scipipe/scipipe@v0.9.10
```
Note: A the time of this writing, v0.9.10 is the latest published version, this wil change over time and you need to adapt

Next, replace any reference to the previous URL with reference to the
<cloned-scipipe-dir> location
```
go mod edit -replace=github.com/scipipe/scipipe@v0.9.10=<cloned-scipipe-dir>
```

Now when you do
```
go run main.go
```
It will not be pulling the code from github but references the code you have cloned locally.

Do your development work and push to your forked repository and do a PR for the author to review and optionally merge the contribution
