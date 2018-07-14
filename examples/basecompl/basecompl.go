package main

import (
	"github.com/scipipe/scipipe"
)

const dna = "AAACCCGGGTTT"

func main() {
	// Initialize workflow
	wf := scipipe.NewWorkflow("DNA Base Complement Workflow", 4)

	// Initialize processes based on shell commands
	makeDNA := wf.NewProc("Make DNA", "echo "+dna+" > {o:seq}")
	baseCompl := wf.NewProc("Base Complement", `cat {i:in} | tr ATCG TAGC > {o:out}`)

	// Connect dependencies
	baseCompl.In("in").From(makeDNA.Out("seq"))

	// Run the workflow
	wf.Run()
}
