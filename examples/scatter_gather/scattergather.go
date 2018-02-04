package main

import (
	"fmt"

	. "github.com/scipipe/scipipe"
	comp "github.com/scipipe/scipipe/components"
)

func main() {
	wf := NewWorkflow("scattergather_wf", 4)

	// === INITIALIZE TASKS =======================================================================

	// Download a zipped Chromosome Y fasta file
	fastaURL := "ftp://ftp.ensembl.org/pub/release-84/fasta/homo_sapiens/dna/Homo_sapiens.GRCh38.dna.chromosome.Y.fa.gz"
	wget := wf.NewProc("wget", "wget "+fastaURL+" -O {o:chry_zipped}")
	wget.SetPathStatic("chry_zipped", "chry.fa.gz")

	// Ungzip the fasta file
	unzip := wf.NewProc("ungzip", "gunzip -c {i:gzipped} > {o:ungzipped}")
	unzip.SetPathReplace("gzipped", "ungzipped", ".gz", "")
	unzip.In("gzipped").Connect(wget.Out("chry_zipped"))

	// Split the fasta file in to parts with 100000 lines in each
	linesPerSplit := 100000
	split := comp.NewFileSplitter(wf, "file_splitter", linesPerSplit)
	split.InFile.Connect(unzip.Out("ungzipped"))

	// Count GC & AT characters in the fasta file
	charCountCommand := "cat {i:infile} | fold -w 1 | grep '[%s]' | wc -l | awk '{ print $1 }' > {o:%s}"
	gccnt := wf.NewProc("gccount", fmt.Sprintf(charCountCommand, "GC", "gccount"))
	gccnt.SetPathExtend("infile", "gccount", ".gccnt")
	gccnt.In("infile").Connect(split.OutSplitFile)

	atcnt := wf.NewProc("atcount", fmt.Sprintf(charCountCommand, "AT", "atcount"))
	atcnt.SetPathExtend("infile", "atcount", ".atcnt")
	atcnt.In("infile").Connect(split.OutSplitFile)

	// Concatenate GC & AT counts
	gccat := comp.NewConcatenator(wf, "gccat", "gccounts.txt")
	gccat.In().Connect(gccnt.Out("gccount"))

	atcat := comp.NewConcatenator(wf, "atcat", "atcounts.txt")
	atcat.In().Connect(atcnt.Out("atcount"))

	// Sum up the GC & AT counts on the concatenated file
	sumCommand := "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}"
	gcsum := wf.NewProc("gcsum", sumCommand)
	gcsum.SetPathExtend("in", "sum", ".sum")
	gcsum.In("in").Connect(gccat.Out())

	atsum := wf.NewProc("atsum", sumCommand)
	atsum.SetPathExtend("in", "sum", ".sum")
	atsum.In("in").Connect(atcat.Out())

	// Finally, calculate the ratio between GC chars, vs. GC+AT chars
	gcrat := wf.NewProc("gcratio", "gc=$(cat {i:gcsum}); at=$(cat {i:atsum}); calc \"$gc/($gc+$at)\" > {o:gcratio}")
	gcrat.SetPathStatic("gcratio", "gcratio.txt")
	gcrat.In("gcsum").Connect(gcsum.Out("sum"))
	gcrat.In("atsum").Connect(atsum.Out("sum"))

	// === RUN PIPELINE ===========================================================================

	wf.Run()
}
