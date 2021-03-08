package scipipe

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestNewProc(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p1 := NewProc(wf, "echo", "echo {p:text}")
	if p1.InParam("text") == nil {
		t.Error(`p.InParam("text") = nil. want: not nil`)
	}

	p2 := NewProc(wf, "cat", "cat {i:infile} > {o:outfile}")
	if p2.In("infile") == nil {
		t.Error(`p.In("infile") = nil. want: not nil`)
	}
	if p2.Out("outfile") == nil {
		t.Error(`p.Out("outfile") = nil. want: not nil`)
	}
}

func TestSetOut(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := wf.NewProc("cat_foo", "cat {i:foo} > {o:bar} # {p:p1}")
	p.InParam("p1").FromStr("p1val")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("data/foo.txt")},
		nil, nil, map[string]string{"p1": "p1val"}, nil, "", nil, 1)

	inputsAndOutputs := map[string]string{
		"{i:foo}":                  "data/foo.txt",
		"{i:foo}.bar.{p:p1}.txt":   "data/foo.txt.bar.p1val.txt",
		"{i:foo|s/.txt//}_bar.txt": "data/foo_bar.txt",
		"{i:foo|s/.txt/.bar/}.txt": "data/foo.bar.txt",
		"{i:foo|%.txt}.bar.txt":    "data/foo.bar.txt", // Bash style strip from end of string
		"{i:foo|%oo.txt}.txt":      "data/f.txt",       // Bash style strip from end of string
		"{i:foo|basename}":         "foo.txt",
	}
	for pathPattern, expectedPath := range inputsAndOutputs {
		// Set a path format for the "bar" out-port
		p.SetOut("bar", pathPattern)

		// Check that the result from running the configured path func (stored in
		// p.PathFuncs) is as expected
		actualPath := p.PathFuncs["bar"](mockTask)
		if actualPath != expectedPath {
			t.Errorf(`Wrong path in SetOut. Got:%v Expected:%v`, actualPath, expectedPath)
		}
	}
}

func TestDefaultPattern(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := wf.NewProc("cat_foo", "cat {i:foo} > {o:bar|.txt} # {p:p1}")
	p.InParam("p1").FromStr("p1val")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, map[string]string{"p1": "p1val"}, nil, "", nil, 1)

	// We expact a filename on the form: input filename . procname . paramname _ val . outport . extension
	expected := "foo.txt.cat_foo.p1_p1val.bar.txt"
	if p.PathFuncs["bar"](mockTask) != expected {
		t.Errorf(`Did not get expected path in SetOut. Got:%v Expected:%v`, p.PathFuncs["bar"](mockTask), expected)
	}
}

func TestDontCreatePortInShellCommand(t *testing.T) {
	wf := NewWorkflow("test_wf", 4)
	ef := wf.NewProc("echo_foo", "echo foo > .tmp/foo.txt")
	ef.SetOut("foo", ".tmp/foo.txt")

	cf := wf.NewProc("cat_foo", "cat {i:foo} > {o:footoo}")
	cf.In("foo").From(ef.Out("foo"))
	cf.SetOut("footoo", ".tmp/footoo.txt")

	wf.Run()

	fileName := ".tmp/footoo.txt"
	f, openErr := os.Open(fileName)
	if openErr != nil {
		t.Errorf("Could not open file: %s\n", fileName)
	}
	b, readErr := ioutil.ReadAll(f)
	if readErr != nil {
		t.Errorf("Could not read file: %s\n", fileName)
	}
	expected := "foo\n"
	if string(b) != expected {
		t.Errorf("File %s did not contain %s as expected, but %s\n", fileName, expected, string(b))
	}

	cleanFiles(".tmp/foo.txt", ".tmp/footoo.txt")
}

func TestProcTaskBuffering(t *testing.T) {
	// An attempt to test the issue in #80.

	numbers := []string{}
	for i := 0; i < getBufsize()+1; i++ {
		numbers = append(numbers, strconv.Itoa(i))
	}

	// Since it's dependent on map iteration order, we run the test
	// multiple times in an attempt to trigger.
	for range [10]int{} {

		wf := NewWorkflow("mwe", 4)

		nSource := NewParamSource(wf, "numbers", numbers...)

		// Make files from params
		stuffMaker := wf.NewProc("make_stuff", "echo testing_{p:number} > {o:out}")
		stuffMaker.InParam("number").From(nSource.Out())
		stuffMaker.SetOut("out", "out/{p:number}.txt")

		// Second Step
		second := wf.NewProc("second_step", "cat {i:in} > {o:out}")
		second.In("in").From(stuffMaker.Out("out"))
		second.SetOut("out", "out/second_{i:in|basename}")

		// Dummy Consumer
		echo := wf.NewProc("echo", "echo {i:first} {i:second}")
		echo.In("first").From(stuffMaker.Out("out"))
		echo.In("second").From(second.Out("out"))

		wf.Run()
	}
}
