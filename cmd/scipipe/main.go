package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	errLog := log.New(os.Stderr, "", 0)
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
	// --------------------------------
	// Create a pipeline runner
	// --------------------------------

	run := sp.NewPipelineRunner()

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	foo := sp.NewProc("fooer",
		"echo foo > {o:foo}")
	foo.SetPathStatic("foo", "foo.txt")
	run.AddProcess(foo)

	f2b := sp.NewProc("foo2bar",
		"sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetPathExtend("foo", "bar", ".bar.txt")
	run.AddProcess(f2b)

	snk := sp.NewSink()
	run.AddProcess(snk)

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------

	f2b.In["foo"].Connect(foo.Out["foo"])
	snk.Connect(f2b.Out["bar"])

	// --------------------------------
	// Run the pipeline!
	// --------------------------------

	run.Run()
}`
				fileName := c.Args().First()
				if fileName == "" {
					fileName = "new_scipipe_workflow.go"
					fmt.Printf("No filename specified, so using the default '%s' ...\n", fileName)
				}

				f, err := os.Create(fileName)
				if err != nil {
					errLog.Println("Could not create file:", fileName)
					os.Exit(1)
				}
				defer f.Close()

				_, err = f.Write([]byte(wfcode))
				if err != nil {
					errLog.Println("Could not write to file:", fileName)
					os.Exit(1)
				}

				fmt.Println("Successfully wrote new workflow file to:", fileName)
				return nil
			},
		},
	}
	app.Run(os.Args)
}
