package main

import (
	"fmt"

	. "github.com/scipipe/scipipe"
	comp "github.com/scipipe/scipipe/components"
)

func main() {
	// === INITIALIZE TASKS =======================================================================

	// Download a zipped Chromosome Y fasta file
	fastaURL := "ftp://ftp.ensembl.org/pub/release-84/fasta/homo_sapiens/dna/Homo_sapiens.GRCh38.dna.chromosome.Y.fa.gz"
	wget := NewFromShell("wget", "wget "+fastaURL+" -O {o:chry_zipped}")
	wget.SetPathStatic("chry_zipped", "chry.fa.gz")

	// Ungzip the fasta file
	unzip := NewFromShell("ungzip", "gunzip -c {i:gzipped} > {o:ungzipped}")
	unzip.SetPathReplace("gzipped", "ungzipped", ".gz", "")

	// Split the fasta file in to parts with 100000 lines in each
	linesPerSplit := 100000
	split := comp.NewFileSplitter("file_splitter", linesPerSplit)

	// Create a 2-way multiplexer that can be used to provide the same
	// file target to two downstream processes
	dupl := comp.NewFanOut("line_duplicator")

	// Count GC & AT characters in the fasta file
	charCountCommand := "cat {i:infile} | fold -w 1 | grep '[%s]' | wc -l | awk '{ print $1 }' > {o:%s}"
	gccnt := NewFromShell("gccount", fmt.Sprintf(charCountCommand, "GC", "gccount"))
	gccnt.SetPathExtend("infile", "gccount", ".gccnt")
	atcnt := NewFromShell("atcount", fmt.Sprintf(charCountCommand, "AT", "atcount"))
	atcnt.SetPathExtend("infile", "atcount", ".atcnt")

	// Concatenate GC & AT counts
	gccat := comp.NewConcatenator("gccat", "gccounts.txt")
	atcat := comp.NewConcatenator("atcat", "atcounts.txt")

	// Sum up the GC & AT counts on the concatenated file
	sumCommand := "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}"
	gcsum := NewFromShell("gcsum", sumCommand)
	gcsum.SetPathExtend("in", "sum", ".sum")
	atsum := NewFromShell("atsum", sumCommand)
	atsum.SetPathExtend("in", "sum", ".sum")

	// Finally, calculate the ratio between GC chars, vs. GC+AT chars
	gcrat := NewFromShell("gcratio", "gc=$(cat {i:gcsum}); at=$(cat {i:atsum}); calc \"$gc/($gc+$at)\" > {o:gcratio}")
	gcrat.SetPathStatic("gcratio", "gcratio.txt")

	// A sink, to drive the network
	asink := NewSink("sink")

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

	wf := NewWorkflow("scattergather_wf")
	wf.AddProcs(wget, unzip, split, dupl, gccnt, atcnt, gccat, atcat, gcsum, atsum, gcrat)
	wf.SetDriver(asink)
	wf.Run()
}
