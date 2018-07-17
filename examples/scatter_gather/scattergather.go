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
	wget.SetOut("chry_zipped", "chry.fa.gz")

	// Ungzip the fasta file
	unzip := wf.NewProc("ungzip", "gunzip -c {i:gz} > {o:ungz}")
	unzip.SetOut("ungz", "{i:gz|%.gz}")
	unzip.In("gz").From(wget.Out("chry_zipped"))

	// Split the fasta file in to parts with 100000 lines in each
	linesPerSplit := 100000
	split := comp.NewFileSplitter(wf, "file_splitter", linesPerSplit)
	split.InFile().From(unzip.Out("ungz"))

	// Count GC & AT characters in the fasta file
	charCountCommand := "cat {i:infile} | fold -w 1 | grep '[%s]' | wc -l | awk '{ print $1 }' > {o:%s}"
	gccnt := wf.NewProc("gccount", fmt.Sprintf(charCountCommand, "GC", "gccount"))
	gccnt.SetOut("gccount", "{i:infile}.gccnt")
	gccnt.In("infile").From(split.OutSplitFile())

	atcnt := wf.NewProc("atcount", fmt.Sprintf(charCountCommand, "AT", "atcount"))
	atcnt.SetOut("atcount", "{i:infile}.atcnt")
	atcnt.In("infile").From(split.OutSplitFile())

	// Concatenate GC & AT counts
	gccat := comp.NewConcatenator(wf, "gccat", "gccounts.txt")
	gccat.In().From(gccnt.Out("gccount"))

	atcat := comp.NewConcatenator(wf, "atcat", "atcounts.txt")
	atcat.In().From(atcnt.Out("atcount"))

	// Sum up the GC & AT counts on the concatenated file
	sumCommand := "awk '{ SUM += $1 } END { print SUM }' {i:in} > {o:sum}"
	gcsum := wf.NewProc("gcsum", sumCommand)
	gcsum.SetOut("sum", "{i:in}.sum")
	gcsum.In("in").From(gccat.Out())

	atsum := wf.NewProc("atsum", sumCommand)
	atsum.SetOut("sum", "{i:in}.sum")
	atsum.In("in").From(atcat.Out())

	// Finally, calculate the ratio between GC chars, vs. GC+AT chars
	gcrat := wf.NewProc("gcratio", "gc=$(cat {i:gcsum}); at=$(cat {i:atsum}); calc \"$gc/($gc+$at)\" > {o:gcratio}")
	gcrat.SetOut("gcratio", "gcratio.txt")
	gcrat.In("gcsum").From(gcsum.Out("sum"))
	gcrat.In("atsum").From(atsum.Out("sum"))

	// === RUN PIPELINE ===========================================================================

	wf.Run()
}
