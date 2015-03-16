# SciPipe

A Go Library for writing Scientific Workflows (so far in pure Go)

This is a work in progress, so more information will come as the
library is developed, but to give a hint about what is coming,
this is how you can write a super-simple workflow already today,
using this library:

```go
package main

import (
    sci "github.com/samuell/scipipe"
)

func main() {
    // Init fooWriter task
    fooWriter := sci.Sh("echo foo > {o:foo1}")
    // Init function for generating output file pattern
    fooWriter.OutPathFuncs["foo1"] = func() string {
        return "foo.txt"
    }

    // Init barReplacer task
    barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo2} > {o:bar}")
    // Init function for generating output file pattern
    barReplacer.OutPathFuncs["bar"] = func() string {
        return barReplacer.GetInPath("foo2") + ".bar"
    }

    // Connect network
    barReplacer.InPorts["foo2"] = fooWriter.OutPorts["foo1"]

    // Set up tasks for execution
    fooWriter.Init()
    barReplacer.Init()

    // Start execution by reading on last port
    <-barReplacer.OutPorts["bar"]
}
```

Executing it will look like so:

```bash
$ ./ex01shell
ShellTask Init(): Executing command:  echo foo > foo.txt
ShellTask Init(): Executing command:  sed 's/foo/bar/g' foo.txt > foo.txt.bar
```

Note especially the way you connect the network - how you do it by just "assigning"
what is in the outport of an upstream task, into the inport of a downstream task.

Then, if inports are not assigned with a channel from an outport,  you can also
manually send targets (filetarget for example) on an inport (they are initialized
with a default channel to make this possible). This could look like so:

```go
package main

import (
    "fmt"
    sci "github.com/samuell/scipipe"
)

func main() {
    // Init barReplacer task
    barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo2} > {o:bar}")
    // Init function for generating output file pattern
    barReplacer.OutPathFuncs["bar"] = func() string {
        return barReplacer.GetInPath("foo2") + ".bar"
    }

    // Set up tasks for execution
    barReplacer.Init()

    // Manually send file targets on the inport of barReplacer
    for _, name := range []string{"foo1", "foo2", "foo3"} {
        barReplacer.InPorts["foo2"] <- sci.NewFileTarget(name + ".txt")
    }
    // We have to manually close the inport as well here, to
    // signal that we are done sending targets (the tasks outport will
    // then automatically be closed as well)
    close(barReplacer.InPorts["foo2"])

    for f := range barReplacer.OutPorts["bar"] {
        fmt.Println("Finished processing file", f.GetPath(), "...")
    }
}
```

Executing this second example might look like so:

```bash
$ ./ex02multifile
ShellTask Init(): Executing command:  sed 's/foo/bar/g' foo1.txt > foo1.txt.bar
ShellTask Init(): Executing command:  sed 's/foo/bar/g' foo2.txt > foo2.txt.bar
Finished processing file foo1.txt.bar ...
ShellTask Init(): Executing command:  sed 's/foo/bar/g' foo3.txt > foo3.txt.bar
Finished processing file foo2.txt.bar ...
Finished processing file foo3.txt.bar ...
```


### Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the near future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
