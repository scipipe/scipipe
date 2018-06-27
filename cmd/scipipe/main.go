package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/scipipe/scipipe"
)

var (
	Info *log.Logger
)

func main() {
	initLogs()
	flag.Parse()
	err := parseFlags(flag.Args())
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
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
	case "audit2html", "audit2htmlflat":
		if len(args) < 2 {
			return errors.New("No infile specified")
		}
		inFile := args[1]
		if len(args) < 3 {
			return errors.New("No outfile specified")
		}
		outFile := args[2]
		err := errors.New("")
		if cmd == "audit2htmlflat" {
			err = auditInfoToHTML(inFile, outFile, true)
		} else {
			err = auditInfoToHTML(inFile, outFile, false)
		}
		if err != nil {
			return errors.Wrap(err, "Could not convert Audit file to HTML")
		}
	default:
		return errors.New("Unknown command: " + cmd)
	}
	return nil
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
	Info.Printf(`________________________________________________________________________

SciPipe (http://scipipe.org)
Version: %s
________________________________________________________________________

Usage:
$ scipipe <command> [command options]

Available commands:
$ scipipe new <filename.go>
$ scipipe audit2html <infile.audit.json> [<outfile.html>]
$ scipipe audit2htmlflat <infile.audit.json> [<outfile.html>]
________________________________________________________________________
`, scipipe.Version)
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

func auditInfoToHTML(inFilePath string, outFilePath string, flatten bool) error {
	ip := scipipe.NewFileIP(strings.Replace(inFilePath, ".audit.json", "", 1))
	auditInfo := ip.AuditInfo()

	outHTML := fmt.Sprintf(`<html><head><style>body { font-family: arial, helvetica, sans-serif; } table { border: 1px solid #ccc; width: 100%%; margin: 1em; } th { text-align: right; vertical-align: top; padding: .2em .8em; width: 140px; } td { vertical-align: top; }</style><title>Audit info for: %s</title></head><body>`, ip.Path())
	if flatten {
		auditInfosByID := extractAuditInfosByID(auditInfo)
		auditInfosByStartTime := sortAuditInfosByStartTime(auditInfosByID)
		for _, ai := range auditInfosByStartTime {
			ai.Upstream = nil
			outHTML += formatTaskHTML(ai.ProcessName, ai)
		}
	} else {
		outHTML += formatTaskHTML(ip.Path(), auditInfo)
	}
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

func formatTaskHTML(fileName string, auditInfo *scipipe.AuditInfo) (outHTML string) {
	outHTML = "<table>\n"
	outHTML += fmt.Sprintf(`<tr><td colspan="2" style="font-size: 1.2em; font-weight: bold; text-align: left; padding: .2em .4em; ">%s</td></tr>`, fileName)
	outHTML += fmt.Sprintf("<tr><th>ID:</th><td>%s</td></tr>\n", auditInfo.ID)
	outHTML += fmt.Sprintf("<tr><th>Process:</th><td>%s</td></tr>\n", auditInfo.ProcessName)
	outHTML += fmt.Sprintf("<tr><th>Command:</th><td><pre>%s</pre></td></tr>\n", auditInfo.Command)

	params := []string{}
	for pname, p := range auditInfo.Params {
		params = append(params, fmt.Sprintf("%s: %s", pname, p))
	}
	outHTML += fmt.Sprintf("<tr><th>Parameters:</th><td>%s</td></tr>\n", strings.Join(params, ", "))
	tags := []string{}
	for pname, p := range auditInfo.Tags {
		tags = append(tags, fmt.Sprintf("%s: %s", pname, p))
	}
	outHTML += fmt.Sprintf("<tr><th>Tags:</th><td><pre>%v</pre></td></tr>\n", strings.Join(tags, ", "))

	outHTML += fmt.Sprintf("<tr><th>Start time:</th><td>%v</td></tr>\n", auditInfo.StartTime)
	outHTML += fmt.Sprintf("<tr><th>Finish time:</th><td>%v</td></tr>\n", auditInfo.FinishTime)
	outHTML += fmt.Sprintf("<tr><th>Execution time:</th><td>%d ms</td></tr>\n", auditInfo.ExecTimeMS)
	upStreamHTML := ""
	for filePath, uai := range auditInfo.Upstream {
		upStreamHTML += formatTaskHTML(filePath, uai)
	}
	if outHTML != "" {
		outHTML += "<tr><th>Upstreams:</th><td>" + upStreamHTML + "</td></tr>\n"
	}
	outHTML += "</table>\n"
	return
}

func extractAuditInfosByID(auditInfo *scipipe.AuditInfo) (auditInfosByID map[string]*scipipe.AuditInfo) {
	auditInfosByID = make(map[string]*scipipe.AuditInfo)
	auditInfosByID[auditInfo.ID] = auditInfo
	for _, ai := range auditInfo.Upstream {
		auditInfosByID = mergeStringAuditInfoMaps(auditInfosByID, extractAuditInfosByID(ai))
	}
	return auditInfosByID
}

func mergeStringAuditInfoMaps(ms ...map[string]*scipipe.AuditInfo) (merged map[string]*scipipe.AuditInfo) {
	merged = make(map[string]*scipipe.AuditInfo)
	for _, m := range ms {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func sortAuditInfosByStartTime(auditInfosByID map[string]*scipipe.AuditInfo) []*scipipe.AuditInfo {
	sorted := []*scipipe.AuditInfo{}

	auditInfosByStartTime := map[time.Time]*scipipe.AuditInfo{}
	startTimes := []time.Time{}
	for _, ai := range auditInfosByID {
		auditInfosByStartTime[ai.StartTime] = ai
		startTimes = append(startTimes, ai.StartTime)
	}
	sort.Slice(startTimes, func(i, j int) bool { return startTimes[i].Before(startTimes[j]) })
	for _, t := range startTimes {
		sorted = append(sorted, auditInfosByStartTime[t])
	}
	return sorted
}
func initLogs() {
	Info = log.New(os.Stdout, "", 0)
}

func initLogsTest() {
	Info = log.New(ioutil.Discard, "", 0)
}

const (
	workflowStub = `// Workflow written in SciPipe.
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
)
