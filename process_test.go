package scipipe

import (
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

func TestSetPathStatic(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := NewProc(wf, "echo_foo", "echo foo > {o:bar}")
	p.SetPathStatic("bar", "bar.txt")

	mockTask := NewTask(wf, p, "echo_foo_task", "", nil, nil, nil, nil, nil, "", nil, 1)

	if p.PathFormatters["bar"](mockTask) != "bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "bar.txt"`)
	}
}

func TestSetPathExtend(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := NewProc(wf, "cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathExtend("foo", "bar", ".bar.txt")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, nil, nil, "", nil, 1)

	if p.PathFormatters["bar"](mockTask) != "foo.txt.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.txt.bar.txt"`)
	}
}

func TestSetPathReplace(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := NewProc(wf, "cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathReplace("foo", "bar", ".txt", ".bar.txt")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, nil, nil, "", nil, 1)

	if p.PathFormatters["bar"](mockTask) != "foo.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.bar.txt"`)
	}
}

func TestSetPathPattern(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := wf.NewProc("cat_foo", "cat {i:foo} > {o:bar} # {p:p1}")
	p.InParam("p1").FromStr("p1val")
	p.SetPathPattern("bar", "{i:foo}.bar.{p:p1}.txt")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, map[string]string{"p1": "p1val"}, nil, "", nil, 1)

	expected := "foo.txt.bar.p1val.txt"
	if p.PathFormatters["bar"](mockTask) != expected {
		t.Errorf(`Did not get expected path in TestSetPathPattern. Got:%v Expected:%v`, p.PathFormatters["bar"](mockTask), expected)
	}
}

func TestDefaultPattern(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := wf.NewProc("cat_foo", "cat {i:foo} > {o:bar.txt} # {p:p1}")
	p.InParam("p1").FromStr("p1val")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, map[string]string{"p1": "p1val"}, nil, "", nil, 1)

	// We expact a filename on the form: input filename . procname . paramname _ val . outport . extension
	expected := "foo.txt.cat_foo.p1_p1val.bar.txt"
	if p.PathFormatters["bar"](mockTask) != expected {
		t.Errorf(`Did not get expected path in TestSetPathPattern. Got:%v Expected:%v`, p.PathFormatters["bar"](mockTask), expected)
	}
}
