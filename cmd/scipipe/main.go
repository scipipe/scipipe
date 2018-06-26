package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scipipe/scipipe"
)

func main() {
	flag.Parse()
	cmd := flag.Arg(0)
	switch cmd {
	case "new":
		writeNewWorkflowFile(flag.Arg(1))
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage of scipipe:\nscipipe new <filename.go>")
}

func writeNewWorkflowFile(fileName string) {
	wfcode := `// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import sp "github.com/scipipe/scipipe"

func main() {
	// Create a workflow, using 4 cpu cores
	wf := sp.NewWorkflow("my_workflow", 4)

	// Initialize processes
	foo := wf.NewProc("fooer", "echo foo > {o:foo}")
	foo.SetPathStatic("foo", "foo.txt")

	f2b := wf.NewProc("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetPathExtend("foo", "bar", ".bar.txt")

	// From workflow dependency network
	f2b.In("foo").From(foo.Out("foo"))

	// Run the workflow
	wf.Run()
}`
	if fileName == "" {
		fmt.Println("ERROR: No filename specified!")
		printUsage()
		os.Exit(1)
	}
	f, err := os.Create(fileName)
	if err != nil {
		scipipe.Fail("Could not create file:", fileName)
	}
	defer f.Close()
	_, err = f.Write([]byte(wfcode))
	if err != nil {
		scipipe.Fail("Could not write to file:", fileName)
	}
	fmt.Println("Successfully wrote new workflow file to:", fileName, "\n\nNow you can run it with:\ngo run ", fileName)
}
