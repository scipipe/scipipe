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

// === References ===
// 1. http://uppnex.se/twiki/do/view/Courses/NgsIntro1502/ResequencingAnalysis
