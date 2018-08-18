There is a component for that,
[FileGlobber](https://godoc.org/github.com/scipipe/scipipe/components#FileGlobber)
(click link for GoDoc API documentation).

A sketchy example, showing how the FileGlobber component can be used, is shown
below:

```go
package main

import (
    "github.com/scipipe/scipipe"
    "github.com/scipipe/scipipe/components"
)

func main() {
    wf := scipipe.NewWorkflow("wf", 4)

    // Initiate a new globber component. Since it is not created from the
    // workflow object, it needs to take the workflow as its first argument,
    // in order to connect itself properly to it.
    globber := components.NewFileGlobber(wf, "globber", "./somedirectory/*")

    // Initiate a command that does some processing on all the globbed files.
    // Extend the command below to do some meaningful processing on all the
    // globbed files
    otherProc := wf.NewProc("otherproc", "cat {i:in} > {o:out}")
    otherProc.In("in").From(globber.Out())

    // Run the workflow
    wf.Run()
}
```

Then, given that you have a number of files in a subdirectory called
`somedirectory`, these should now be captured by the globbing component, and
sent to `otherProc` for processing.
