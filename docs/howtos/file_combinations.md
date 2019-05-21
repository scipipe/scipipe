Sometimes you need to create all the possible combinations of a set of files
that you have as file streams. This you can do with the
[FileCombinator](https://godoc.org/github.com/scipipe/scipipe/components#FileCombinator)
component.

## Example

Given that you have a set of files:

```
letterfile_a.txt
letterfile_b.txt
numberfile_1.txt
numberfile_2.txt
numberfile_3.txt
```

... and you want to create all combinations of the letter files and the number
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

This will generate the following files (excluding the accompanying .audit.json files):

```
letterfile_a.txt
letterfile_a.numberfile_1.combined.txt
letterfile_a.numberfile_2.combined.txt
letterfile_a.numberfile_3.combined.txt
letterfile_b.txt
letterfile_b.numberfile_1.combined.txt
letterfile_b.numberfile_2.combined.txt
letterfile_b.numberfile_3.combined.txt
```
