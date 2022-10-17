package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"errors"

	"github.com/scipipe/scipipe"
)

var (
	Info *log.Logger
)

func main() {
	initLogs()
	initHelp()
	flag.Parse()
	err := parseFlags(flag.Args())
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func parseFlags(args []string) error {
	if len(args) < 1 {
		printHelp()
		return nil
	}
	cmd := args[0]
	switch cmd {
	case "new":
		if len(args) < 2 {
			return errors.New("No infile specified")
		}
		writeNewWorkflowFile(args[1])
	case "audit2html":
		inFile, outFile, err := parseArgsAudit2X(args, "html")
		if err != nil {
			return errWrap(err, "Could not parse filenames from arguments")
		}
		err = auditInfoToHTML(inFile, outFile, true)
		if err != nil {
			return errWrap(err, "Could not convert Audit file to HTML")
		}
	case "audit2tex":
		inFile, outFile, err := parseArgsAudit2X(args, "tex")
		if err != nil {
			return errWrap(err, "Could not parse filenames from arguments")
		}
		err = auditInfoToTeX(inFile, outFile, true)
		if err != nil {
			return errWrap(err, "Could not convert Audit file to TeX")
		}
	case "audit2bash":
		inFile, outFile, err := parseArgsAudit2X(args, "sh")
		if err != nil {
			return errWrap(err, "Could not parse filenames from arguments")
		}
		err = auditInfoToBash(inFile, outFile, true)
		if err != nil {
			return errWrap(err, "Could not convert Audit file to Bash")
		}
	default:
		return errors.New("Unknown command: " + cmd)
	}
	return nil
}

func parseArgsAudit2X(args []string, extension string) (inFile string, outFile string, err error) {
	if len(args) < 2 {
		return "", "", errors.New("No infile specified")
	}
	inFile = args[1]
	if len(inFile) < 12 || (inFile[len(inFile)-11:] != ".audit.json") {
		return "", "", errors.New("Infile does not look like an audit file (does not end with .audit.json): " + inFile)
	}
	outFile = strings.Replace(inFile, ".audit.json", ".audit."+extension, 1)

	if len(args) > 3 {
		return "", "", errors.New("Extra arguments found: " + args[3])
	}
	if len(args) == 3 {
		outFile = args[2]
	}
	return
}

func printNewUsage() {
	Info.Println(`
Usage:
$ scipipe new <filename.go>`)
}
func printAudit2HTMLUsage() {
	Info.Print(`
Usage:
$ scipipe audit2html <infile.audit.json> [<outfile.html>]
`)
}

func printHelp() {
	flag.Usage()
}

func writeNewWorkflowFile(fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		scipipe.Fail("Could not create file:", fileName)
	}
	defer f.Close()
	_, err = f.Write([]byte(workflowStub))
	if err != nil {
		scipipe.Fail("Could not write to file:", fileName)
	}
	Info.Println("Successfully wrote new workflow file to:", fileName, "\n\nNow you can run it with:\ngo run ", fileName)
}

func initHelp() {
	flag.Usage = func() {
		fmt.Printf(`________________________________________________________________________

SciPipe (http://scipipe.org)
Version: %s
________________________________________________________________________

Usage:
$ scipipe <command> [command options]

Available commands:
$ scipipe new <filename.go>
$ scipipe audit2html <infile.audit.json> [<outfile.html>]
$ scipipe audit2tex <infile.audit.json> [<outfile.tex>]
$ scipipe audit2bash <infile.audit.json> [<outfile.sh>]
________________________________________________________________________
`, scipipe.Version)
	}
}

func initLogs() {
	Info = log.New(os.Stdout, "", 0)
}

func initLogsTest() {
	Info = log.New(ioutil.Discard, "", 0)
}

const workflowStub = `// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import sp "github.com/scipipe/scipipe"

func main() {
	// Create a workflow, using 4 cpu cores
	wf := sp.NewWorkflow("my_workflow", 4)

	// Initialize processes
	foo := wf.NewProc("fooer", "echo foo > {o:foo}")
	foo.SetOut("foo", "foo.txt")

	f2b := wf.NewProc("foo2bar", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	f2b.SetOut("bar", "{i:foo}.bar.txt")

	// From workflow dependency network
	f2b.In("foo").From(foo.Out("foo"))

	// Run the workflow
	wf.Run()
}`

func errWrap(err error, msg string) error {
	return errors.New(msg + "\nOriginal error: " + err.Error())
}

func errWrapf(err error, msg string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, v...) + "\nOriginal error: " + err.Error())
}
