package main

import (
	"fmt"

	sp "github.com/scipipe/scipipe"
	"github.com/scipipe/scipipe/components"
)

func main() {
	// === INITIALIZE TASKS =======================================================================

	// Download a zipped Chromosome Y fasta file
	fastaURL := "ftp://ftp.ensembl.org/pub/release-84/fasta/homo_sapiens/dna/Homo_sapiens.GRCh38.dna.chromosome.Y.fa.gz"
	wget := sp.NewFromShell("wget", "wget "+fastaURL+" -O {o:chry_zipped}")
	wget.SetPathStatic("chry_zipped", "chry.fa.gz")

	// Ungzip the fasta file
	unzip := sp.NewFromShell("ungzip", "gunzip -c {i:gzipped} > {o:ungzipped}")
	unzip.SetPathReplace("gzipped", "ungzipped", ".gz", "")

	// Split the fasta file in to parts with 100000 lines in each
	linesPerSplit := 100000
	split := components.NewFileSplitter(linesPerSplit)

	// Create a 2-way multiplexer that can be used to provide the same
	// file target to two downstream processes
	dupl := components.NewFanOut()

	// Count GC & AT characters in the fasta file
	charCountCommand := "cat {i:infile} | fold -w 1 | grep '[%s]' | wc -l | awk '{ print $1 }' > {o:%s}"
	gccnt := sp.NewFromShell("gccount", fmt.Sprintf(charCountCommand, "GC", "gccount"))
	gccnt.SetPathExtend("infile", "gccount", ".gccnt")
	atcnt := sp.NewFromShell("atcount", fmt.Sprintf(charCountCommand, "AT", "atcount"))
	atcnt.SetPathExtend("infile", "atcount", ".atcnt")

	// Concatenate GC & AT counts
	gccat := components.NewConcatenator("gccounts.txt")
	atcat := components.NewConcatenator("atcounts.txt")

	// Sum up the GC & AT counts on the concatenated file
	sumCommand := "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}"
	gcsum := sp.NewFromShell("gcsum", sumCommand)
	gcsum.SetPathExtend("in", "sum", ".sum")
	atsum := sp.NewFromShell("atsum", sumCommand)
	atsum.SetPathExtend("in", "sum", ".sum")

	// Finally, calculate the ratio between GC chars, vs. GC+AT chars
	gcrat := sp.NewFromShell("gcratio", "gc=$(cat {i:gcsum}); at=$(cat {i:atsum}); calc \"$gc/($gc+$at)\" > {o:gcratio}")
	gcrat.SetPathStatic("gcratio", "gcratio.txt")

	// A sink, to drive the network
	asink := sp.NewSink()

	// === CONNECT DEPENDENCIES ===================================================================

	unzip.In("gzipped").Connect(wget.Out("chry_zipped"))
	split.InFile.Connect(unzip.Out("ungzipped"))
	dupl.InFile.Connect(split.OutSplitFile)
	gccnt.In("infile").Connect(dupl.Out("gccnt"))
	atcnt.In("infile").Connect(dupl.Out("atcnt"))
	gccat.In.Connect(gccnt.Out("gccount"))
	atcat.In.Connect(atcnt.Out("atcount"))
	gcsum.In("in").Connect(gccat.Out)
	atsum.In("in").Connect(atcat.Out)
	gcrat.In("gcsum").Connect(gcsum.Out("sum"))
	gcrat.In("atsum").Connect(atsum.Out("sum"))

	asink.Connect(gcrat.Out("gcratio"))

	// === RUN PIPELINE ===========================================================================

	piperunner := sp.NewPipelineRunner()
	piperunner.AddProcesses(wget, unzip, split, dupl, gccnt, atcnt, gccat, atcat, gcsum, atsum, gcrat, asink)
	piperunner.Run()
}
