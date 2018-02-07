SciPipe allows you to, on-demand, run only specific parts of a workflow. This
can be useful especially if you are doing modifications far up in an already
developed workflow, and want to run only up to a specific process, rather
than also running all downstream processes, which might be unnecessary heavy.

This can be done by using the
[workflow.RunTo()](https://godoc.org/github.com/scipipe/scipipe#Workflow.RunTo)
method. By using this instead of the normal `workflow.Run()` method, scipipe
will only run this process and all upstream processes of that one.

See also a
[simple&nbsp;example](https://github.com/scipipe/scipipe/blob/master/examples/run_specific_procs/run_specific_procs.go)
of where this is used.

There are a few other variants for specifying parts of workflows (and more
might be added in the future), such as specifying individual process names,
or providing the process structs themselves. Please refer to the relevant
parts of the
[workflow&nbsp;documentation](https://godoc.org/github.com/scipipe/scipipe#Workflow)
for more about that.
