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
	fastaURL := "ftp://ftp.ensembl.org/pub/release-84/fasta/homo_sapiens/dna/Homo_sapiens.GRCh38.dna.chromosome.Y.fa.gz"
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
	gccat := scipipe.NewConcatenator("gccounts.txt")
	gcsum := scipipe.Shell("gcsum", "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}")
	gcsum.SetPathFormatExtend("in", "sum", ".sum")

	atcnt := scipipe.Shell("atcount", "cat {i:infile} | fold -w 1 | grep '[AT]' | wc -l | awk '{ print $1 }' > {o:atcount}")
	atcnt.SetPathFormatExtend("infile", "atcount", ".atcnt")
	atcat := scipipe.NewConcatenator("atcounts.txt")
	atsum := scipipe.Shell("atsum", "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}")
	atsum.SetPathFormatExtend("in", "sum", ".sum")

	gcrat := scipipe.Shell("gcratio", "gc=$(cat {i:gcsum}); at=$(cat {i:atsum}); calc \"$gc/($gc+$at)\" > {o:gcratio}")
	gcrat.SetPathFormatStatic("gcratio", "gcratio.txt")

	asink := scipipe.NewSink()

	// ------------------------------------------------------------------------
	// CONNECT DEPENDENCIES
	// ------------------------------------------------------------------------

	unzip.InPorts["gzipped"] = wget.OutPorts["chry_zipped"]
	split.InFile = unzip.OutPorts["ungzipped"]

	dupl.InFile = split.OutSplitFile

	gccnt.InPorts["infile"] = dupl.OutFile1
	atcnt.InPorts["infile"] = dupl.OutFile2

	gccat.In = gccnt.OutPorts["gccount"]
	atcat.In = atcnt.OutPorts["atcount"]

	gcsum.InPorts["in"] = gccat.Out
	atsum.InPorts["in"] = atcat.Out

	gcrat.InPorts["gcsum"] = gcsum.OutPorts["sum"]
	gcrat.InPorts["atsum"] = atsum.OutPorts["sum"]

	asink.In = gcrat.OutPorts["gcratio"]

	piperunner := scipipe.NewPipelineRunner()
	piperunner.AddProcesses(wget, unzip, split, dupl, gccnt, atcnt, gccat, atcat, gcsum, atsum, gcrat, asink)

	// ------------------------------------------------------------------------
	// RUN PIPELINE
	// ------------------------------------------------------------------------

	piperunner.Run()
}
