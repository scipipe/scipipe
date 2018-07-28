[SciPipe 0.8.0](https://github.com/scipipe/scipipe/releases/tag/v0.8.0)
introduced a feature to plot a directed graph of workflows in SciPipe [1].
This can be done in two ways:

1. Just producing a DOT text file, with the graph definition
2. Also converting this DOT file to PDF.

Number 1. above can be done without any external dependencies, while number 2
requires that graphviz, with the `dot` command is installed on the system (On
Ubuntu it can be installed with the command: `sudo apt-get install graphviz`).

## How to plot graphs

To write a .dot file in SciPipe, include a line like follows, in your workflow
definition, provided that you have initiated the variable `wf` with a workflow
struct:

```go
func main() {
    wf := scipipe.NewWorkflow("my workflow", 4)
    // Workflow code here
    wf.PlotGraph("my_workflow_graph.dot") // <-- SEE THIS LINE!
    wf.Run()
}
```

If you want to also convert the dot file to PDF in one go, instead change the
next last line to:

```go
    wf.PlotGraphPDF("my_workflow_graph.dot")
```

## How to plot graphs conditionally based on a flag

Now, you might not want to generate a new plot every time you run your workflow
(although, perhaps you would? ... checking in a .dot version of your workflow
could in fact be a great way to keep a more readable version of your workflow
at hand ... but anyhow), you could make the plotting optional, based on a flag.
This is something we've found ourselves doing quite often at pharmb.io. This
could be done as follows (more complete code example):

```go
package main

import (
    "flag"
    "github.com/scipipe/scipipe"
)

var (
    plotGraph = flag.Bool("plotgraph", false, "Plot a directed graph of the workflow to PDF")
)

func main() {
    flag.Parse()

    wf := scipipe.NewWorkflow("testwf", 4)
    wf.NewProc("foo", "echo foo > {o:out}")

    if *plotGraph {
        wf.PlotGraphPDF("wfgraph.dot")
    }
    wf.Run()
}
```

Now, the graph will only plotted if you run your workflow with the
`-plotgraph` flag, e.g:

```bash
go run myworkflow.go -plotgraph
```

## Links

- [GoDoc for Workflow.PlotGraph()](https://godoc.org/github.com/scipipe/scipipe#Workflow.PlotGraph)
- [GoDoc for Workflow.PlotGraphPDF()](https://godoc.org/github.com/scipipe/scipipe#Workflow.PlotGraphPDF)

## Footnotes

[1] these are often called "DAG" for "Directed Acyclic Graph", but
SciPipe does not have a guarantee or requirement on acyclicness of the graph,
thus just "directed graph".