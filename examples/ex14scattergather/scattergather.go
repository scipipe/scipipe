package main

import (
	"github.com/samuell/scipipe"
)

func main() {
	//scipipe.InitLogDebug()

	// ------------------------------------------------------------------------
	// INITIALIZE TASKS
	// ------------------------------------------------------------------------

	// Download a zipped Chromosome Y fasta file
	fastaURL := "ftp://ftp.ensembl.org/pub/release-67/fasta/homo_sapiens/dna/Homo_sapiens.GRCh37.67.dna_rm.chromosome.Y.fa.gz"
	wget := scipipe.Shell("wget", "wget "+fastaURL+" -O {o:chry_zipped}")
	wget.SetPathFormatStatic("chry_zipped", "chry.fa.gz")

	// Ungzip the fasta file
	unzip := scipipe.Shell("ungzip", "gunzip -c {i:gzipped} > {o:ungzipped}")
	unzip.SetPathFormatReplace("gzipped", "ungzipped", ".gz", "")

	// Split the fasta file in to parts with 100000 lines in each
	linesPerSplit := 100000
	split := scipipe.NewFileSplitter(linesPerSplit)

	// Create a 2-way multiplexer that can be used to provide the same
	// file target to two downstream processes
	dupl := scipipe.NewDuplicator()

	// Count lines in the fasta file
	gccnt := scipipe.Shell("gccount", "cat {i:infile} | fold -w 1 | grep '[GC]' | wc -l | awk '{ print $1 }' > {o:gccount}")
	gccnt.SetPathFormatExtend("infile", "gccount", ".gccnt")

	atcnt := scipipe.Shell("atcount", "cat {i:infile} | fold -w 1 | grep '[AT]' | wc -l | awk '{ print $1 }' > {o:atcount}")
	atcnt.SetPathFormatExtend("infile", "atcount", ".atcnt")

	merge := scipipe.NewMerger()

	asink := scipipe.NewSink()

	// ------------------------------------------------------------------------
	// CONNECT DEPENDENCIES
	// ------------------------------------------------------------------------

	unzip.InPorts["gzipped"] = wget.OutPorts["chry_zipped"]
	split.InFile = unzip.OutPorts["ungzipped"]

	dupl.InFile = split.OutSplitFile

	gccnt.InPorts["infile"] = dupl.OutFile1
	atcnt.InPorts["infile"] = dupl.OutFile2

	merge.Ins = append(merge.Ins, gccnt.OutPorts["gccount"])
	merge.Ins = append(merge.Ins, atcnt.OutPorts["atcount"])

	asink.In = merge.Out

	piperunner := scipipe.NewPipelineRunner()
	piperunner.AddProcesses(wget, unzip, split, dupl, gccnt, atcnt, merge, asink)

	// ------------------------------------------------------------------------
	// RUN PIPELINE
	// ------------------------------------------------------------------------

	piperunner.Run()
}
