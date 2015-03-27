# SciPipe

A [Go(lang)](http://golang.org) Library for writing Scientific Workflows (so far in pure Go)

This library is an experiment in building a scientific workflow engine in pure Go,
based on an idea for a flow-based like pattern in pure Go, as presented by the author in
[this blog post on Gopher Academy](http://blog.gopheracademy.com/composable-pipelines-pattern).

From flow-based programming, It uses the principles of separate network (workflow dependency graph)
definition, named in- and out-ports and bounded buffers (already available in Go) to make
writing workflows as easy as possible.

In addition to that, it adds convenience methods (see `sp.Sh()` below) for creating ad hoc tasks
on the fly, using a shell command, with inputs and outputs defined in-line in the shell command,
with a syntax of `{i:INPORT_NAME}` for inports, and `{o:OUTPORT_NAME}` for outports.o

Example: Creating two example tasks:

```go
fooWriter := sp.Sh("echo 'foo' > {o:outfile}")
fooToBarReplacer := sp.Sh("cat {i:foofile} | sed 's/foo/bar/g' > {o:barfile}")
```

For these inports and outports, channels for sending and receiving FileTargets are automatically
created and put in a hashmap added as a struct field of the task, named `InPorts` and `OutPorts` repectively,
Eash channel is added to the hashmap with its inport/outport name as key in the hashmap,
so that the channel can be retrieved from the hashmap using the in/outport name.

Connecting outports of one task to the inport of another task is then done by assigning the
respective channels to the corresponding places in the hashmap.

Example: Connecting the two tasks creating above:

```go
fooToBarReplacer.InPorts["foofile"] = fooWriter.OutPorts["outfile"]
```

The only thing remaining after this, is to provide some way for the program to figure out a
suitable file name for each of the files propagating through this little "network" of tasks.
This is done by adding a closure (function) to another special hashmap, again keyed by
the names of the outports of the tasks. So, to define the output filenames of the two tasks
above, we would add:

```go
fooWriter.OutPathFuncs["outfile"] = func() string {
	// Just statically create a file named foo.txt
	return "foo.txt"
}
fooToBarReplacer.OutPathFuncs["barfile"] = func() string {
	// Here, we instead re-use the file name of the task we depend
	// on (which we get on the 'foofile' inport), and just
	// pad '.bar' at the end:
	return fooToBarReplacer.GetInPath("foofile") + ".bar"
}
```

So with this, we have done everything needed to set up a file-based batch workflow system:

- Specified task dependencies by wiring outputs of the upstream tasks to inports in downstream tasks.
- For each outport, provided a function that will compute a suitable file name for the new file.

For a complete, more real-world example, see the code here below, which shows how you can write a simple
bioinformatics (sequence alignment) workflow already today, using this library,
implementing a few steps of an [NGS bioinformatics tutorial](uppnex.se/twiki/do/view/Courses/NgsIntro1502/ResequencingAnalysis)
held at [SciLifeLab](http://www.scilifelab.se) in Uppsala in February 2015:

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

    // In this specific case we send two inputs on the same port,
    // basically meaning that the align task will run twice,
    // producing two outputs:
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

- This library is heavily influenced/inspired by (and might make use of on in the future),
  the [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- It is also heavily influenced by the [Flow-based programming](http://www.jpaulmorrison.com/fbp) by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
- This work is financed by faculty grants and other financing for Jarl Wikberg's [Pharmaceutical Bioinformatics group](http://www.farmbio.uu.se/forskning/researchgroups/pb/) of Dept. of
  Pharmaceutical Biosciences at Uppsala University. Main supervisor for the project is [Ola Spjuth](http://www.farmbio.uu.se/research/researchgroups/pb/olaspjuth).
