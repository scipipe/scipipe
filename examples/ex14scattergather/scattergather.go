package main

import (
	"fmt"

	"github.com/scipipe/scipipe"
)

func main() {
	// === INITIALIZE TASKS =======================================================================

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
	dupl := scipipe.NewFanOut()

	// Count GC & AT characters in the fasta file
	charCountCommand := "cat {i:infile} | fold -w 1 | grep '[%s]' | wc -l | awk '{ print $1 }' > {o:%s}"
	gccnt := scipipe.Shell("gccount", fmt.Sprintf(charCountCommand, "GC", "gccount"))
	gccnt.SetPathFormatExtend("infile", "gccount", ".gccnt")
	atcnt := scipipe.Shell("atcount", fmt.Sprintf(charCountCommand, "AT", "atcount"))
	atcnt.SetPathFormatExtend("infile", "atcount", ".atcnt")

	// Concatenate GC & AT counts
	gccat := scipipe.NewConcatenator("gccounts.txt")
	atcat := scipipe.NewConcatenator("atcounts.txt")

	// Sum up the GC & AT counts on the concatenated file
	sumCommand := "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}"
	gcsum := scipipe.Shell("gcsum", sumCommand)
	gcsum.SetPathFormatExtend("in", "sum", ".sum")
	atsum := scipipe.Shell("atsum", sumCommand)
	atsum.SetPathFormatExtend("in", "sum", ".sum")

	// Finally, calculate the ratio between GC chars, vs. GC+AT chars
	gcrat := scipipe.Shell("gcratio", "gc=$(cat {i:gcsum}); at=$(cat {i:atsum}); calc \"$gc/($gc+$at)\" > {o:gcratio}")
	gcrat.SetPathFormatStatic("gcratio", "gcratio.txt")

	// A sink, to drive the network
	asink := scipipe.NewSink()

	// === CONNECT DEPENDENCIES ===================================================================

	unzip.InPorts["gzipped"] = wget.OutPorts["chry_zipped"]
	split.InFile = unzip.OutPorts["ungzipped"]
	dupl.InFile = split.OutSplitFile
	gccnt.InPorts["infile"] = dupl.GetOutPort("gccnt")
	atcnt.InPorts["infile"] = dupl.GetOutPort("atcnt")
	gccat.In = gccnt.OutPorts["gccount"]
	atcat.In = atcnt.OutPorts["atcount"]
	gcsum.InPorts["in"] = gccat.Out
	atsum.InPorts["in"] = atcat.Out
	gcrat.InPorts["gcsum"] = gcsum.OutPorts["sum"]
	gcrat.InPorts["atsum"] = atsum.OutPorts["sum"]
	asink.In = gcrat.OutPorts["gcratio"]

	// === RUN PIPELINE ===========================================================================

	piperunner := scipipe.NewPipelineRunner()
	piperunner.AddProcesses(wget, unzip, split, dupl, gccnt, atcnt, gccat, atcat, gcsum, atsum, gcrat, asink)
	piperunner.Run()
}
