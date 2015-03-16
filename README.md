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

### Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the near future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
