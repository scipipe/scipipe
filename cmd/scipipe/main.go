package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/pkg/errors"
	"github.com/scipipe/scipipe"
)

var (
	Info *log.Logger
)

func main() {
	flag.Parse()
	err := parseFlags(flag.Args())
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
	}
}

func parseFlags(args []string) error {
	scipipe.InitLogError()
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
			return errors.Wrap(err, "Could not parse filenames from arguments")
		}
		err = auditInfoToHTML(inFile, outFile, true)
		if err != nil {
			return errors.Wrap(err, "Could not convert Audit file to HTML")
		}
	case "audit2tex":
		inFile, outFile, err := parseArgsAudit2X(args, "tex")
		if err != nil {
			return errors.Wrap(err, "Could not parse filenames from arguments")
		}
		err = auditInfoToTeX(inFile, outFile, true)
		if err != nil {
			return errors.Wrap(err, "Could not convert Audit file to TeX")
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
	Info.Printf(`________________________________________________________________________

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

	outHTML := fmt.Sprintf(headHTMLPattern, ip.Path())
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
	outHTML += bottomHTML

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
	outHTML += fmt.Sprintf("	<tr><td colspan=\"2\" class=\"task-title\"><strong>%s</strong> / <a name=\"%s\" href=\"#%s\"><code>%s</code></a></td></tr>\n", auditInfo.ProcessName, auditInfo.ID, auditInfo.ID, auditInfo.ID)
	outHTML += fmt.Sprintf("	<tr><td colspan=\"2\"><div class=\"cmdbox\">%s</div></td></tr>\n", auditInfo.Command)

	params := []string{}
	for pname, p := range auditInfo.Params {
		params = append(params, fmt.Sprintf("%s: %s", pname, p))
	}
	outHTML += fmt.Sprintf("	<tr><th>Parameters:</th><td>%s</td></tr>\n", strings.Join(params, ", "))
	tags := []string{}
	for pname, p := range auditInfo.Tags {
		tags = append(tags, fmt.Sprintf("%s: %s", pname, p))
	}
	outHTML += fmt.Sprintf("	<tr><th>Tags:</th><td><pre>%v</pre></td></tr>\n", strings.Join(tags, ", "))

	outHTML += fmt.Sprintf("	<tr><th>Start time:</th><td>%s</td></tr>\n", auditInfo.StartTime.Format(`2006-01-02 15:04:05<span class="greyout">.000 -0700 MST</span>`))
	outHTML += fmt.Sprintf("	<tr><th>Finish time:</th><td>%s</td></tr>\n", auditInfo.FinishTime.Format(`2006-01-02 15:04:05<span class="greyout">.000 -0700 MST</span>`))
	et := auditInfo.ExecTimeNS
	outHTML += fmt.Sprintf("	<tr><th>Execution time:</th><td>%s</td></tr>\n", et.Truncate(time.Millisecond).String())
	//upStreamHTML := ""
	//for filePath, uai := range auditInfo.Upstream {
	//	upStreamHTML += formatTaskHTML(filePath, uai)
	//}
	//if outHTML != "" {
	//	outHTML += "<tr><th>Upstreams:</th><td>" + upStreamHTML + "</td></tr>\n"
	//}
	outHTML += "</table>\n"
	return
}

func auditInfoToTeX(inFilePath string, outFilePath string, flatten bool) error {
	outFile, err := os.Create(outFilePath)
	scipipe.CheckWithMsg(err, "Could not create TeX file")

	auditInfo := scipipe.UnmarshalAuditInfoJSONFile(inFilePath)
	auditInfosByID := extractAuditInfosByID(auditInfo)
	auditInfosByStartTime := sortAuditInfosByStartTime(auditInfosByID)

	texTpl := template.New("TeX").Funcs(
		template.FuncMap{
			"strrepl":        func(subj string, find string, repl string) string { return strings.Replace(subj, find, repl, -1) },
			"sub":            func(val1 int, val2 int) int { return val1 - val2 },
			"timesub":        func(t1 time.Time, t2 time.Time) time.Duration { return t1.Sub(t2) },
			"durtomillis":    func(exact time.Duration) (rounded time.Duration) { return exact.Truncate(1e6 * time.Nanosecond) },
			"timetomillis":   func(exact time.Time) (rounded time.Time) { return exact.Truncate(1e6 * time.Nanosecond) },
			"durtomillisint": func(exact time.Duration) (millis int) { return int(exact.Nanoseconds() / 1000000) },
		})
	texTpl, err = texTpl.Parse(texTemplate)
	scipipe.CheckWithMsg(err, "Could not parse TeX template")

	report := auditReport{
		FileName:    inFilePath,
		ScipipeVer:  scipipe.Version,
		RunTime:     auditInfosByStartTime[len(auditInfosByStartTime)-1].FinishTime.Sub(auditInfosByStartTime[0].StartTime),
		AuditInfos:  auditInfosByStartTime,
		ChartHeight: fmt.Sprintf("%.03f", 1.0+float64(len(auditInfosByStartTime))*0.6),
	}

	palette, err1 := colorful.WarmPalette(len(report.AuditInfos))
	if err1 != nil {
		scipipe.CheckWithMsg(err1, "Could not create color palette")
	}
	for i, col := range palette {
		r, g, b := col.RGB255()
		report.ColorDef += fmt.Sprintf("\\definecolor{color%d}{RGB}{%d,%d,%d}\n", i, r, g, b)
	}

	texTpl.Execute(outFile, report)
	return nil
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
	headHTMLPattern = `<html>
<head>
<style>
	body { font-family: arial, helvetica, sans-serif; }
	table { color: #546E7A; background: #EFF2F5; border: none; width: 960px; margin: 1em 1em 2em 1em; padding: 1.2em; font-size: 10pt; opacity: 1; }
	table:hover { color: black; background: #FFFFEF; }
	th { text-align: right; vertical-align: top; padding: .2em .8em; width: 9em; }
	td { vertical-align: top; }
	.task-title { font-size: 12pt; font-weight: normal; }
	.cmdbox { border: rgb(156, 184, 197) 0px solid; background: #D2DBE0; font-family: 'Ubuntu mono', Monospace, 'Courier New'; padding: .8em 1em; margin: 0.4em 0; font-size: 12pt; }
	table:hover .cmdbox { background: #EFEFCC; }
	.greyout { color: #999; }
	a, a:link, a:visited { color: inherit; text-decoration: none; }
	a:hover { text-decoration: underline; }
</style>
<title>Audit info for: %s</title>
</head>
<body>
`
	bottomHTML = `</body>
</html>`
)

// LaTeX code from vision.tex:
const texTemplate = `\documentclass[11pt,oneside,openright]{memoir}

\usepackage{tcolorbox}
\usepackage[scaled]{beramono}
\renewcommand*\familydefault{\ttdefault}
\usepackage[T1]{fontenc}
\usepackage{tabularx}
\usepackage{listings}
\usepackage{graphicx}
\usepackage{tikz}
\usepackage{pgfplots}
\usepackage{pgfplotstable}
\usepackage{xcolor}

{{ .ColorDef }}

% from https://tex.stackexchange.com/a/128040/110842
% filter to only get the current row in \pgfplotsinvokeforeach
\pgfplotsset{
    select row/.style={
        x filter/.code={\ifnum\coordindex=#1\else\def\pgfmathresult{}\fi}
    }
}

\pgfplotstableread[col sep=comma]{
start,end,Name,color
{{ $startTime := (index .AuditInfos 0).StartTime }}
{{ range $i, $v := .AuditInfos }}{{ durtomillisint (timesub $v.StartTime $startTime) }},{{ durtomillisint (timesub $v.FinishTime $startTime) }},{{ strrepl .ProcessName "_" "\\_" }},color{{ $i }}
{{ end }}
}\loadedtable
\pgfplotstablegetrowsof{\loadedtable}
\pgfplotsset{compat=1.13}
\pgfmathsetmacro{\tablerows}{int(\pgfplotsretval-1)}

\begin{document}
\pagestyle{plain}
\noindent
\begin{minipage}{\textwidth}
    \vspace{-8em}\hspace{-8em}
    %\includegraphics[width=9em]{images/scipipe_logo_bluegrey.png}
\end{minipage}

\noindent
{\huge\textbf{SciPipe Audit Report}} \\
\vspace{10pt}

    \begin{tcolorbox}[ title=Workflow for file: {{ (strrepl (strrepl .FileName ".audit.json" "") "_" "\\_") }} ]
    \small
\begin{tabular}{rp{0.72\linewidth}}
SciPipe version: & {{ .ScipipeVer }} \\
Start time:  & {{ timetomillis (index .AuditInfos 0).StartTime }} \\
Finish time: & {{ timetomillis (index .AuditInfos (sub (len .AuditInfos) 1)).FinishTime }} \\
Run time: & {{ durtomillis .RunTime }}  \\
\end{tabular}
    \end{tcolorbox}

\setlength{\fboxsep}{0pt}
\noindent

%\hspace{-0.1725\textwidth}\fbox{\includegraphics[width=1.35\textwidth]{images/cawpre.pdf}}

\section*{Execution timeline}

\begin{tikzpicture}
\begin{axis}[
    xbar, xmin=0,
    y axis line style = { opacity = 0 },
    tickwidth         = 0pt,
	width=10cm,
	height={{ .ChartHeight }}cm,
    % next two lines also from https://tex.stackexchange.com/a/128040/110842,
    ytick={0,...,\tablerows},
    yticklabels from table={\loadedtable}{Name},
    xbar stacked,
    bar shift=0pt,
    y dir=reverse,
    xtick={1, 60000, 120000, 180000, 240000, 300000, 600000, 900000, 1200000},
    xticklabels={0, 1 min, 2 min, 3 min, 4 min, 5 min, 10 min, 15 min, 20 min},
    scaled x ticks=false,
]

\pgfplotsinvokeforeach{0,...,\tablerows}{
    % get color from table, commands defined must be individual for each plot
    % because the color is used in \end{axis} and therefore would otherwise
    % use the last definition
    \pgfplotstablegetelem{#1}{color}\of{\loadedtable}
    \expandafter\edef\csname barcolor.#1\endcsname{\pgfplotsretval}
    \addplot+[color=\csname barcolor.#1\endcsname] table [select row=#1, x expr=\thisrow{end}-\thisrow{start}, y expr=#1]{\loadedtable};
}
\end{axis}
\end{tikzpicture}

\newpage

\section*{Tasks}
    \lstset{ breaklines=true,
            postbreak=\mbox{\textcolor{red}{$\hookrightarrow$}\space},
            aboveskip=8pt,belowskip=8pt}

{{ range $i, $v := .AuditInfos }}
   \begin{tcolorbox}[ title={{ (strrepl $v.ProcessName "_" "\\_") }},
                      colbacktitle=color{{ $i }}!63!white,
                      colback=color{{ $i }}!37!white,
                      coltitle=black ]
       \small
       \begin{tabular}{rp{0.72\linewidth}}
ID: & {{ $v.ID }} \\
Process: & {{ (strrepl $v.ProcessName "_" "\\_") }} \\
Command: & \begin{lstlisting}
{{ strrepl $v.Command "_" "\\_" }}
\end{lstlisting} \\
Parameters:& {{ range $k, $v := $v.Params }}{{- $k -}}={{- $v -}}{{ end }} \\
Tags: & {{ range $k, $v := $v.Tags }}{{- $k -}}={{- $v -}}{{ end }} \\
Start time:  & {{ timetomillis $v.StartTime }} \\
Finish time: & {{ timetomillis $v.FinishTime }} \\
Execution time: & {{ durtomillis $v.ExecTimeNS }} \\
        \end{tabular}
	\end{tcolorbox}
{{ end }}

\end{document}`
