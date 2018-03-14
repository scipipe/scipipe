*Beware: Technical topic, best suited for power-users!*

If you want to write a component with Go code, but would like to have it work
seamlessly with other workflow processes in SciPipe, without reimplementing the
whole [Process](https://godoc.org/github.com/scipipe/scipipe#Process)
functionality yourself, there is a way to do it: By using the `CustomExecute`
field of Process.

In short, it can be done like this:

```go
// Initiate task from a "shell like" pattern, though here we
// just specify the out-port, and nothing else. We have to
// specify the out-port (and any other ports we plan to use later),
// so that they are correctly initialized.
fooWriter := sci.NewProc("fooer", "{o:foo}")

// Set the output formatter to a static string
fooWriter.SetPathStatic("foo", "foo.txt")

// Create the custom execute function, with pure Go code and
// add it to the CustomExecute field of the fooWriter process
fooWriter.CustomExecute = func(task *sci.Task) {
    task.OutIP("foo").Write([]byte("foo\n"))
}
```

For a more detailed example, see [this example](https://github.com/scipipe/scipipe/blob/master/examples/custom_execution_function/funchook.go)
(Have a look at the [NewFooer()](https://github.com/scipipe/scipipe/blob/master/examples/custom_execution_function/funchook.go#L34-L50)
and [NewFoo2Barer()](https://github.com/scipipe/scipipe/blob/master/examples/custom_execution_function/funchook.go#L72-L89)
factory functions in particular!)
