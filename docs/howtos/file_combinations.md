Sometimes you need to create all the possible combinations of a set of files
that you have as file streams. 

For example, say that you have two file streams:

```
[a.txt b.txt]
[1.txt 2.txt 3.txt]
```

... and you want to process all of the combinations of these two sets of files.
So in other words, what you want is:

```
[a.txt a.txt a.txt b.txt b.txt b.txt]
[1.txt 2.txt 3.txt 1.txt 2.txt 3.txt]
```

This is something you can accomplish with the [FileCombinator](https://godoc.org/github.com/scipipe/scipipe/components#FileCombinator)
component, available in [SciPipe 0.9.1](https://github.com/scipipe/scipipe/releases/tag/v0.9.1)
and later.

## Example

Given that you have a set of files:

```
letterfile_a.txt
letterfile_b.txt
numberfile_1.txt
numberfile_2.txt
numberfile_3.txt
```

... and you want to create all combinations of the `letter*` files and the `number*`
files, you can do it as follows:


```go
package main

import (
    "github.com/scipipe/scipipe"
    "github.com/scipipe/scipipe/components"
)

func main() {
    wf := scipipe.NewWorkflow("wf", 4)

    letterGlobber := components.NewFileGlobber(wf, "letter_globber", "letterfile_*.txt")
    numberGlobber := components.NewFileGlobber(wf, "number_globber", "numberfile_*.txt")

    fileCombiner := components.NewFileCombinator(wf, "file_combiner")
    fileCombiner.In("letters").From(letterGlobber.Out())
    fileCombiner.In("numbers").From(numberGlobber.Out())

    catenator := wf.NewProc("catenator", "cat {i:letters} {i:numbers} > {o:combined}")
    catenator.In("letters").From(fileCombiner.Out("letters"))
    catenator.In("numbers").From(fileCombiner.Out("numbers"))
    catenator.SetOut("combined", "{i:letters|basename|%.txt}.{i:numbers|basename|%.txt}.combined.txt")

    wf.Run()
}
```

Note that when accessing an in-port on the FileCombinator with the `In(PORTNAME)` method, this port
will be created automatically, together with a corresponding out-port which can be accessed with the
same name, `Out(PORTNAME)`, as can be seen when we connect the fileCombinator to the catenator process
further down in the code.

The program above, if put in a `.go` file and run with `go run file.go`, will generate the following
files (excluding the accompanying .audit.json files):

```
letterfile_b.txt
letterfile_a.txt
numberfile_3.txt
numberfile_2.txt
numberfile_1.txt
letterfile_a.numberfile_2.combined.txt
letterfile_a.numberfile_1.combined.txt
letterfile_a.numberfile_3.combined.txt
letterfile_b.numberfile_2.combined.txt
letterfile_b.numberfile_1.combined.txt
letterfile_b.numberfile_3.combined.txt
```

As you can see, all the combinations of the 
