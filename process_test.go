package scipipe

import (
	"testing"
)

func TestNewProc(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p1 := NewProc(wf, "echo", "echo {p:text}")
	if p1.ParamInPort("text") == nil {
		t.Error(`p.ParamInPort("text") = nil. want: not nil`)
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

	mockTask := NewTask(wf, p, "echo_foo_task", "", nil, nil, nil, nil, "", nil, 1)

	if p.PathFormatters["bar"](mockTask) != "bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "bar.txt"`)
	}
}

func TestSetPathExtend(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := NewProc(wf, "cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathExtend("foo", "bar", ".bar.txt")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, nil, "", nil, 1)

	if p.PathFormatters["bar"](mockTask) != "foo.txt.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.txt.bar.txt"`)
	}
}

func TestSetPathReplace(t *testing.T) {
	wf := NewWorkflow("test_wf", 16)
	p := NewProc(wf, "cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathReplace("foo", "bar", ".txt", ".bar.txt")

	mockTask := NewTask(wf, p, "echo_foo_task", "", map[string]*FileIP{"foo": NewFileIP("foo.txt")}, nil, nil, nil, "", nil, 1)

	if p.PathFormatters["bar"](mockTask) != "foo.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.bar.txt"`)
	}
}
