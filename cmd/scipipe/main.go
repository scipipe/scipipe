package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/scipipe/scipipe"
)

func main() {
	scipipe.InitLogError()

	flag.Parse()
	cmd := flag.Arg(0)
	switch cmd {
	case "new":
		if len(flag.Args()) < 2 {
			fmt.Println("ERROR: No filename specified!")
			printNewUsage()
			os.Exit(1)
		}
		writeNewWorkflowFile(flag.Arg(1))
	case "audit2html":
		if len(flag.Args()) < 2 {
			fmt.Println("ERROR: No infile specified!")
			printAudit2HTMLUsage()
			os.Exit(1)
		}
		inFile := flag.Arg(1)
		if len(flag.Args()) < 3 {
			fmt.Println("ERROR: No outfile specified!")
			printAudit2HTMLUsage()
			os.Exit(1)
		}
		outFile := flag.Arg(2)
		err := convertAudit2Html(inFile, outFile)
		if err != nil {
			scipipe.CheckWithMsg(err, "Could not convert Audit file to HTML")
		}
	default:
		printHelp()
	}
}

func printNewUsage() {
	fmt.Print(`
Usage:
$ scipipe new <filename.go>
`)
}
func printAudit2HTMLUsage() {
	fmt.Print(`
Usage:
$ scipipe audit2html <infile.audit.json> [<outfile.html>]
`)
}

func printHelp() {
	fmt.Printf(`________________________________________________________________________

SciPipe (http://scipipe.org)
Version: %s
________________________________________________________________________

Usage:
$ scipipe <command> [command options]

Available commands:
$ scipipe new <filename.go>
$ scipipe audit2html <infile.audit.json> [<outfile.html>]
________________________________________________________________________
`, scipipe.Version)
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

func convertAudit2Html(inFilePath string, outFilePath string) error {
	ip := scipipe.NewFileIP(strings.Replace(inFilePath, ".audit.json", "", 1))
	ai := ip.AuditInfo()

	outHTML := fmt.Sprintf(`<html><head><style>body { font-family: arial, helvetica, sans-serif; } table { border: 1px solid #ccc; } th { text-align: right; vertical-align: top; padding: .2em .8em; } td { vertical-align: top; }</style><title>Audit info for: %s</title></head><body>`, ip.Path())
	outHTML += formatAuditInfoHTML(ip.Path(), ai)
	outHTML += `</body></html>`

	if _, err := os.Stat(outFilePath); os.IsExist(err) {
		return errors.Wrap(err, "File already exists:"+outFilePath)
	}
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return errors.Wrap(err, "Could not create file:"+outFilePath)
	}
	outFile.WriteString(outHTML)
	outFile.Close()
	return nil
}

func formatAuditInfoHTML(fileName string, ai *scipipe.AuditInfo) (outHTML string) {
	outHTML = "<table>\n"
	outHTML += fmt.Sprintf(`<tr><td colspan="2" style="font-size: 1.2em; font-weight: bold; text-align: left; padding: .2em .4em; ">%s</td></tr>`, fileName)
	outHTML += fmt.Sprintf("<tr><th>ID:</th><td>%s</td></tr>\n", ai.ID)
	outHTML += fmt.Sprintf("<tr><th>Process:</th><td>%s</td></tr>\n", ai.ProcessName)
	outHTML += fmt.Sprintf("<tr><th>Command:</th><td><pre>%s</pre></td></tr>\n", ai.Command)

	params := []string{}
	for pname, p := range ai.Params {
		params = append(params, fmt.Sprintf("%s: %s", pname, p))
	}
	outHTML += fmt.Sprintf("<tr><th>Parameters:</th><td>%s</td></tr>\n", strings.Join(params, ", "))
	tags := []string{}
	for pname, p := range ai.Tags {
		tags = append(tags, fmt.Sprintf("%s: %s", pname, p))
	}
	outHTML += fmt.Sprintf("<tr><th>Tags:</th><td><pre>%v</pre></td></tr>\n", strings.Join(tags, ", "))

	outHTML += fmt.Sprintf("<tr><th>Start time:</th><td>%v</td></tr>\n", ai.StartTime)
	outHTML += fmt.Sprintf("<tr><th>Finish time:</th><td>%v</td></tr>\n", ai.FinishTime)
	outHTML += fmt.Sprintf("<tr><th>Execution time:</th><td>%d ms</td></tr>\n", ai.ExecTimeMS)
	upStreamHTML := ""
	for filePath, uai := range ai.Upstream {
		upStreamHTML += formatAuditInfoHTML(filePath, uai)
	}
	if outHTML != "" {
		outHTML += "<tr><th>Upstreams:</th><td>" + upStreamHTML + "</td></tr>\n"
	}
	outHTML += "</table>\n"
	return
}
