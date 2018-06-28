package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/scipipe/scipipe"
)

func TestNewCmd(t *testing.T) {
	initLogsTest()

	testWfPath := "/tmp/testwf.go"

	args := []string{"new", testWfPath}
	err := parseFlags(args)
	if err != nil {
		t.Error("Could not parse flags:", err.Error())
	}

	if _, err := os.Stat(testWfPath); os.IsNotExist(err) {
		t.Error(t, "`scipipe new` command failed to create new workflow file: "+testWfPath)
	}

	cleanFiles(t, testWfPath)
}

func TestAudit2HTMLCmd(t *testing.T) {
	initLogsTest()

	jsonFile := "/tmp/f.audit.json"
	htmlFile := "/tmp/f.audit.html"
	err := ioutil.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Error("Could not create infile needed in test: " + jsonFile)
	}

	// Test both audit2html commands
	for cmd, expectedContent := range map[string]string{
		"audit2html": htmlContent,
	} {
		args := []string{cmd, jsonFile, htmlFile}
		err = parseFlags(args)
		if err != nil {
			t.Error("Could not parse flags:", err.Error())
		}
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			t.Error("`scipipe audit2html` command failed to create HTML file: " + htmlFile)
		}
		htmlBytes, err := ioutil.ReadFile(htmlFile)
		if err != nil {
			t.Error(errors.Wrap(err, "Could not read HTML file:"+htmlFile).Error())
		}
		if string(htmlBytes) != expectedContent {
			t.Errorf("Converted HTML content of %s was not as expected.\nEXPECTED:\n%s\nACTUAL:\n%s\n", htmlFile, expectedContent, string(htmlBytes))
		}
		cleanFiles(t, htmlFile)
	}
	cleanFiles(t, jsonFile)
}

func TestExtractAuditInfosByID(t *testing.T) {
	ai1 := scipipe.NewAuditInfo()
	ai1Cmd := "echo foo > foo.txt"
	ai1.Command = ai1Cmd

	ai2 := scipipe.NewAuditInfo()
	ai2Cmd := "sed 's/foo/bar/g' foo.txt > bar.txt"
	ai2.Command = ai2Cmd
	ai2.Upstream["foo.txt"] = ai1

	aiByID := extractAuditInfosByID(ai2)
	if aiByID == nil {
		t.Errorf("Extracted audit info by ID map is nil: %v", aiByID)
	}
	aiByIDExpectedLength := 2
	if len(aiByID) != aiByIDExpectedLength {
		t.Errorf("Extracted audit info by ID map has wrong length.\nExpected: %d, Actual: %d\n", aiByIDExpectedLength, len(aiByID))
	}
	for _, ai := range aiByID {
		if ai.Command != ai1Cmd && ai.Command != ai2Cmd {
			t.Errorf("Command in extracted audit info by ID was wrong.\nExpected: '%s' OR '%s'\nActual: '%s'\n", ai1Cmd, ai2Cmd, ai.Command)
		}
	}
}

const (
	jsonContent = `{
	"ID": "omlcgx0izet4bprr7e5f",
	"ProcessName": "foo2bar",
	"Command": "sed 's/foo/bar/g' ../fooer.foo.txt \u003e fooer.foo.txt.foo2bar.bar.txt",
	"Params": {},
	"Tags": {},
	"StartTime": "2018-06-27T17:50:51.445311702+02:00",
	"FinishTime": "2018-06-27T17:50:51.451388569+02:00",
	"ExecTimeNS": 6000000,
	"Upstream": {
		"fooer.foo.txt": {
			"ID": "y23kkipm4p4y7kgdzuc1",
			"ProcessName": "fooer",
			"Command": "echo foo \u003e fooer.foo.txt",
			"Params": {},
			"Tags": {},
			"StartTime": "2018-06-27T17:50:51.437331897+02:00",
			"FinishTime": "2018-06-27T17:50:51.44444825+02:00",
			"ExecTimeNS": 7000000,
			"Upstream": {}
		}
	}
}`

	htmlContent = `<html>
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
<title>Audit info for: /tmp/f</title>
</head>
<body>
<table>
	<tr><td colspan="2" class="task-title"><strong>fooer</strong> / <a name="y23kkipm4p4y7kgdzuc1" href="#y23kkipm4p4y7kgdzuc1"><code>y23kkipm4p4y7kgdzuc1</code></a></td></tr>
	<tr><td colspan="2"><div class="cmdbox">echo foo > fooer.foo.txt</div></td></tr>
	<tr><th>Parameters:</th><td></td></tr>
	<tr><th>Tags:</th><td><pre></pre></td></tr>
	<tr><th>Start time:</th><td>2018-06-27 17:50:51<span class="greyout">.437 +0200 CEST</span></td></tr>
	<tr><th>Finish time:</th><td>2018-06-27 17:50:51<span class="greyout">.444 +0200 CEST</span></td></tr>
	<tr><th>Execution time:</th><td>7ms</td></tr>
</table>
<table>
	<tr><td colspan="2" class="task-title"><strong>foo2bar</strong> / <a name="omlcgx0izet4bprr7e5f" href="#omlcgx0izet4bprr7e5f"><code>omlcgx0izet4bprr7e5f</code></a></td></tr>
	<tr><td colspan="2"><div class="cmdbox">sed 's/foo/bar/g' ../fooer.foo.txt > fooer.foo.txt.foo2bar.bar.txt</div></td></tr>
	<tr><th>Parameters:</th><td></td></tr>
	<tr><th>Tags:</th><td><pre></pre></td></tr>
	<tr><th>Start time:</th><td>2018-06-27 17:50:51<span class="greyout">.445 +0200 CEST</span></td></tr>
	<tr><th>Finish time:</th><td>2018-06-27 17:50:51<span class="greyout">.451 +0200 CEST</span></td></tr>
	<tr><th>Execution time:</th><td>6ms</td></tr>
</table>
</body>
</html>`
)

func cleanFiles(t *testing.T, files ...string) {
	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			t.Error(err.Error())
		}
	}
}
