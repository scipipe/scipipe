# SciPipe

A Go Library for writing Scientific Workflows (so far in pure Go)

This is a work in progress, so more information will come as the
library is developed, but to give a hint about what is coming,
this is how you can write a simple NGS (sequence alignment)
bioinformatics workflow already today, using this library:

```go
package main

import (
    "fmt"
    sp "github.com/samuell/scipipe"
    re "regexp"
)

const (
    REF      = "human_17_v37.fasta"
    BASENAME = ".ILLUMINA.low_coverage.17q_"
)

var (
    INDIVIDUALS = [2]string{"NA06984", "NA07000"}
    SAMPLES     = [2]string{"1", "2"}
)

func main() {
    // Initialize existing files
    fastq1 := sp.NewFileTarget(fmt.Sprintf("%s%s1.fq", INDIVIDUALS[0], BASENAME))
    fastq2 := sp.NewFileTarget(fmt.Sprintf("%s%s2.fq", INDIVIDUALS[1], BASENAME))

    // Step 2 in [1]
    align := sp.Sh("bwa aln " + REF + " {i:fastq} > {o:sai}")
    align.OutPathFuncs["sai"] = func() string {
        return align.GetInPath("fastq") + ".sai"
    }

    // Step 3 in [1]
    merge := sp.Sh("bwa sampe " + REF + " {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged}")
    merge.OutPathFuncs["merged"] = func() string {
        ptrn, err := re.Compile("NA[0-9]+")
        sp.Check(err)
        ind1 := ptrn.FindString(merge.GetInPath("sai1"))
        ind2 := ptrn.FindString(merge.GetInPath("sai2"))
        return ind1 + "." + ind2 + ".merged.sam"
    }

    // Wire the dataflow network / dependency graph
    merge.InPorts["sai1"] = align.OutPorts["sai"]
    merge.InPorts["sai2"] = align.OutPorts["sai"]

    // For some of the inputs, we just send file targets "manually"
    // (where they don't come from a previous task)
    align.InPorts["fastq"] <- fastq1
    align.InPorts["fastq"] <- fastq2

    merge.InPorts["fq1"] <- fastq1
    merge.InPorts["fq2"] <- fastq2

    // Set up tasks for execution
    align.Init()
    merge.Init()

    // Run pipeline by asking for the final output
    <-merge.OutPorts["merged"]
}
```

### Acknowledgements

- This library is heavily influenced/inspired by (and might make use of on in the near future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
