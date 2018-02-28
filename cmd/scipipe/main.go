package main

import (
	"fmt"
	"os"

	"github.com/scipipe/scipipe"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "The SciPipe Tool"
	app.Usage = "A helper tool to ease working with SciPipe workflows"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "Create a new example workflow file, in a filename provided as first argument after 'new' (default: mynewworkflow.go).",
			Action: func(c *cli.Context) error {
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

	// Connect workflow dependency network
	f2b.In("foo").Connect(foo.Out("foo"))

	// Run the workflow
	wf.Run()
}`
				fileName := c.Args().First()
				if fileName == "" {
					fileName = "my_workflow.go"
					fmt.Printf("No filename specified, so using the default '%s' ...\n", fileName)
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

				fmt.Println("Successfully wrote new workflow file to:", fileName)
				return nil
			},
		},
	}
	app.Run(os.Args)
}
