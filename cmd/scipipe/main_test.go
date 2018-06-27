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

	jsonFile := "/tmp/fooer.foo.txt.foo2bar.bar.txt.audit.json"
	htmlFile := "/tmp/fooer.foo.txt.foo2bar.bar.txt.audit.html"

	err := ioutil.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Error("Could not create infile needed in test: " + jsonFile)
	}

	args := []string{"audit2html", jsonFile, htmlFile}
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

	if string(htmlBytes) != htmlContent {
		t.Errorf("Converted HTML content of %s was not as expected.\nEXPECTED:\n%s\nACTUAL:\n%s\n", htmlFile, htmlContent, string(htmlBytes))
	}

	//cleanFiles(t, jsonFile, htmlFile)
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
	"ExecTimeMS": 6,
	"Upstream": {
		"fooer.foo.txt": {
			"ID": "y23kkipm4p4y7kgdzuc1",
			"ProcessName": "fooer",
			"Command": "echo foo \u003e fooer.foo.txt",
			"Params": {},
			"Tags": {},
			"StartTime": "2018-06-27T17:50:51.437331897+02:00",
			"FinishTime": "2018-06-27T17:50:51.44444825+02:00",
			"ExecTimeMS": 7,
			"Upstream": {}
		}
	}
}`
	htmlContent = `<html><head><style>body { font-family: arial, helvetica, sans-serif; } table { border: 1px solid #ccc; width: 100%; margin: 1em; } th { text-align: right; vertical-align: top; padding: .2em .8em; width: 140px; } td { vertical-align: top; }</style><title>Audit info for: /tmp/fooer.foo.txt.foo2bar.bar.txt</title></head><body><table>
<tr><td colspan="2" style="font-size: 1.2em; font-weight: bold; text-align: left; padding: .2em .4em; ">/tmp/fooer.foo.txt.foo2bar.bar.txt</td></tr><tr><th>ID:</th><td>omlcgx0izet4bprr7e5f</td></tr>
<tr><th>Process:</th><td>foo2bar</td></tr>
<tr><th>Command:</th><td><pre>sed 's/foo/bar/g' ../fooer.foo.txt > fooer.foo.txt.foo2bar.bar.txt</pre></td></tr>
<tr><th>Parameters:</th><td></td></tr>
<tr><th>Tags:</th><td><pre></pre></td></tr>
<tr><th>Start time:</th><td>2018-06-27 17:50:51.445311702 +0200 CEST</td></tr>
<tr><th>Finish time:</th><td>2018-06-27 17:50:51.451388569 +0200 CEST</td></tr>
<tr><th>Execution time:</th><td>6 ms</td></tr>
<tr><th>Upstreams:</th><td><table>
<tr><td colspan="2" style="font-size: 1.2em; font-weight: bold; text-align: left; padding: .2em .4em; ">fooer.foo.txt</td></tr><tr><th>ID:</th><td>y23kkipm4p4y7kgdzuc1</td></tr>
<tr><th>Process:</th><td>fooer</td></tr>
<tr><th>Command:</th><td><pre>echo foo > fooer.foo.txt</pre></td></tr>
<tr><th>Parameters:</th><td></td></tr>
<tr><th>Tags:</th><td><pre></pre></td></tr>
<tr><th>Start time:</th><td>2018-06-27 17:50:51.437331897 +0200 CEST</td></tr>
<tr><th>Finish time:</th><td>2018-06-27 17:50:51.44444825 +0200 CEST</td></tr>
<tr><th>Execution time:</th><td>7 ms</td></tr>
<tr><th>Upstreams:</th><td></td></tr>
</table>
</td></tr>
</table>
</body></html>`
)

func cleanFiles(t *testing.T, files ...string) {
	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			t.Error(err.Error())
		}
	}
}
