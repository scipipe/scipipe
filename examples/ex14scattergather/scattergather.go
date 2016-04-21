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
	fasta_url := "ftp://ftp.ensembl.org/pub/release-67/fasta/homo_sapiens/dna/Homo_sapiens.GRCh37.67.dna_rm.chromosome.Y.fa.gz"
	wget := scipipe.Shell("wget", "wget "+fasta_url+" -O {o:chry_zipped}")
	wget.SetPathFormatStatic("chry_zipped", "chry.fa.gz")

	// Ungzip the fasta file
	unzip := scipipe.Shell("ungzip", "gunzip -c {i:gzipped} > {o:ungzipped}")
	unzip.SetPathFormatReplace("gzipped", "ungzipped", ".gz", "")

	// Split the fasta file in to parts with 100000 lines in each
	linesPerSplit := 100000
	split := scipipe.NewFileSplitter(linesPerSplit)

	// Count lines in the fasta file
	lncnt := scipipe.Shell("linecount", "wc -l {i:infile} | awk '{ print $1 }' > {o:linecount}")
	lncnt.SetPathFormatExtend("infile", "linecount", ".linecnt")

	asink := scipipe.NewSink()

	// ------------------------------------------------------------------------
	// CONNECT DEPENDENCIES
	// ------------------------------------------------------------------------

	unzip.InPorts["gzipped"] = wget.OutPorts["chry_zipped"]
	split.InFile = unzip.OutPorts["ungzipped"]
	lncnt.InPorts["infile"] = split.OutSplitFile
	asink.In = lncnt.OutPorts["linecount"]

	piperunner := scipipe.NewPipelineRunner()
	piperunner.AddProcesses(wget, unzip, split, lncnt, asink)

	// ------------------------------------------------------------------------
	// RUN PIPELINE
	// ------------------------------------------------------------------------

	piperunner.Run()
}
