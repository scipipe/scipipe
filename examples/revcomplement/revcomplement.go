package main

import (
	"github.com/scipipe/scipipe"
)

const dna = "AAAGCCCGTGGGGGACCTGTTC"

func main() {
	// Initialize workflow, using max 4 CPU cores
	wf := scipipe.NewWorkflow("DNA Base Complement Workflow", 4)

	// Initialize processes based on shell commands:

	// makeDNA writes a DNA string to a file
	makeDNA := wf.NewProc("Make DNA", "echo "+dna+" > {o:dna}")
	makeDNA.SetOut("dna", "dna.txt")

	// complmt computes the base complement of a DNA string
	complmt := wf.NewProc("Base Complement", "cat {i:in} | tr ATCG TAGC > {o:compl}")
	complmt.SetOut("compl", "{i:in|%.txt}.compl.txt")

	// reverse reverses the input DNA string
	reverse := wf.NewProc("Reverse", "cat {i:in} | rev > {o:rev}")
	reverse.SetOut("rev", "{i:in|%.txt}.rev.txt")

	// Connect data dependencies between out- and in-ports
	complmt.In("in").From(makeDNA.Out("dna"))
	reverse.In("in").From(complmt.Out("compl"))

	// Run the workflow
	wf.Run()
}
