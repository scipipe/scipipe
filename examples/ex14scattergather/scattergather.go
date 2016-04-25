package main

import (
	"github.com/samuell/scipipe"
)

func main() {
	scipipe.InitLogAudit()

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
	mul2x := scipipe.NewMultiplexerX2()

	// Count lines in the fasta file
	gccnt := scipipe.Shell("gccount", "cat {i:infile} | fold -w 1 | grep '[GC]' | wc -l | awk '{ print $1 }' > {o:gccount}")
	gccnt.SetPathFormatExtend("infile", "gccount", ".gccnt")

	atcnt := scipipe.Shell("atcount", "cat {i:infile} | fold -w 1 | grep '[AT]' | wc -l | awk '{ print $1 }' > {o:atcount}")
	atcnt.SetPathFormatExtend("infile", "atcount", ".atcnt")

	asink := scipipe.NewSink()

	// ------------------------------------------------------------------------
	// CONNECT DEPENDENCIES
	// ------------------------------------------------------------------------

	unzip.InPorts["gzipped"] = wget.OutPorts["chry_zipped"]
	split.InFile = unzip.OutPorts["ungzipped"]

	mul2x.InFile = split.OutSplitFile

	gccnt.InPorts["infile"] = mul2x.OutFile1
	atcnt.InPorts["infile"] = mul2x.OutFile2

	// Here we have to do the assignment the other way: From the sink to the
	// upstream tasks, so that both of the upstream tasks point to the same
	// downstream channel.
	gccnt.OutPorts["gccount"] = asink.In
	atcnt.OutPorts["atcount"] = asink.In

	piperunner := scipipe.NewPipelineRunner()
	piperunner.AddProcesses(wget, unzip, split, mul2x, gccnt, atcnt, asink)

	// ------------------------------------------------------------------------
	// RUN PIPELINE
	// ------------------------------------------------------------------------

	piperunner.Run()
}
