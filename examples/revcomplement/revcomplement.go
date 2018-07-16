package main

import (
	"github.com/scipipe/scipipe"
)

const dna = "AAACCCGGGTTT"

func main() {
	// Initialize workflow
	wf := scipipe.NewWorkflow("DNA Base Complement Workflow", 4)

	// Initialize processes based on shell commands
	makeDNA := wf.NewProc("Make DNA", "echo "+dna+" > {o:dna}")
	complmt := wf.NewProc("Base Complement", `cat {i:in} | tr ATCG TAGC > {o:compl}`)
	reverse := wf.NewProc("Reverse", "cat {i:in} | rev > {o:rev}")

	// Connect dependencies
	complmt.In("in").From(makeDNA.Out("dna"))
	reverse.In("in").From(complmt.Out("compl"))

	// Run the workflow
	wf.Run()
}
