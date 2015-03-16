package main

import (
	"fmt"
	sp "github.com/samuell/scipipe"
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
	//for i in INDIVIDUALS:
	//	# Workflow definition
	//	# ---------------------------------------------------------------------------------
	//	# files() will return a pseudo task that just outputs an existing file,
	//	#         while not running anything.
	//	# shell() will create a new task with a command that can take inputs
	//	#         and outputs.
	// Initialize existing files
	fastq1 := sp.NewFileTarget(fmt.Sprintf("%s%s1.fq", INDIVIDUALS[0], BASENAME))
	fastq2 := sp.NewFileTarget(fmt.Sprintf("%s%s2.fq", INDIVIDUALS[1], BASENAME))
	//	fq1 = file('fastq:{i}/{i}{b}1.fq'.format(i=i,b=BASENAME))
	//	fq2 = file('fastq:{i}/{i}{b}2.fq'.format(i=i,b=BASENAME))
	//
	//	# Step 2 in [1]--------------------------------------------------------------------
	//	aln1 = shell('bwa aln {ref} <i:fastq> > <o:sai:<i:fastq>.sai>'.format(ref=REF))
	align := sp.Sh("bwa aln " + REF + " {i:fastq} > {o:sai}")
	align.OutPathFuncs["sai"] = func() string {
		return align.GetInPath("fastq") + ".sai"
	}
	//	aln1.inports['fastq'] = fq1.outport('fastq')
	//
	//	aln2 = shell('bwa aln {ref} <i:fastq> > <o:sai:<i:fastq>.sai>'.format(ref=REF))
	//	aln2.inports['fastq'] = fq2.outport('fastq')
	//
	//	# Step 3 in [1]--------------------------------------------------------------------
	//	merg = shell('bwa sampe {ref} <i:sai1> <i:sai2> <i:fq1> <i:fq2> > <o:merged:{i}/{i}{b}.merged.sam>'.format(
	merge := sp.Sh("bwa sampe " + REF + " {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged}")
	merge.OutPathFuncs["merged"] = func() string {
		return merge.GetInPath("sai1") + "." + merge.GetInPath("sai2") + ".merged.sam"
	}
	//		ref=REF,
	//		i=i,
	//		b=BASENAME))
	//	merg.inports['sai1'] = aln1.outport('sai')
	//	merg.inports['sai2'] = aln2.outport('sai')
	//	merg.inports['fq1'] = fq1.outport('fastq')
	//	merg.inports['fq2'] = fq2.outport('fastq')

	align.Init()
	merge.Init()

	merge.InPorts["sai1"] = align.OutPorts["sai"]
	merge.InPorts["sai2"] = align.OutPorts["sai"]

	// Wire the network
	align.InPorts["fastq"] <- fastq1
	align.InPorts["fastq"] <- fastq2
	//close(align.InPorts["fastq"])

	merge.InPorts["fq1"] <- fastq1
	//close(merge.InPorts["fq1"])
	merge.InPorts["fq2"] <- fastq2
	//close(merge.InPorts["fq2"])

	<-merge.OutPorts["merged"]
}
