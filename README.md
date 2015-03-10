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
	barReplacer := sci.Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo") + ".bar"
	}
	barReplacer.Init()

	for _, name := range []string{"foo1.txt", "foo2.txt", "foo3.txt"} {
		barReplacer.InPorts["foo"] <- sci.NewFileTarget(name)
	}
	close(barReplacer.InPorts["foo"])

	for {
		<-barReplacer.OutPorts["bar"]
	}
}
```

### Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the near future), [GoFlow](https://github.com/trustmaster/goflow).
- It is alsy heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John-Paul Morrison](http://www.jpaulmorrison.com/fbp).
