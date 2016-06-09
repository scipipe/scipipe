package scipipe

import (
	"testing"
)

func TestNewFromShell(t *testing.T) {
	p1 := NewFromShell("echo", "echo {p:text}")
	if p1.ParamPorts["text"] == nil {
		t.Error(`p.ParamPorts["text"] = nil. want: not nil`)
	}

	p2 := NewFromShell("cat", "cat {i:infile} > {o:outfile}")
	if p2.InPorts["infile"] == nil {
		t.Error(`p.OutPorts["infile"] = nil. want: not nil`)
	}
	if p2.OutPorts["outfile"] == nil {
		t.Error(`p.OutPorts["outfile"] = nil. want: not nil`)
	}
}

func TestShellExpand_OnlyParams(t *testing.T) {
	p1 := ShellExpand("echo", "echo {p:foo}", nil, nil, map[string]string{"foo": "bar"})
	if p1.CommandPattern != "echo bar" {
		t.Error(`p.CommandPattern != "echo bar", want: echo bar`)
	}
}

func TestShellExpand_InputOutput(t *testing.T) {
	p := ShellExpand("cat", "cat {i:foo} > {o:bar}", map[string]string{"foo": "foo.txt"}, map[string]string{"bar": "bar.txt"}, nil)
	if p.CommandPattern != "cat foo.txt > bar.txt" {
		t.Error(`if p.CommandPattern != "cat foo.txt > bar.txt"`)
	}
}

func TestSetPathStatic(t *testing.T) {
	p := NewFromShell("echo_foo", "echo foo > {o:bar}")
	p.SetPathStatic("bar", "bar.txt")

	mock_task := NewSciTask("echo_foo_task", "", nil, nil, nil, nil, "")

	if p.PathFormatters["bar"](mock_task) != "bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "bar.txt"`)
	}
}

func TestSetPathExtend(t *testing.T) {
	p := NewFromShell("cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathExtend("foo", "bar", ".bar.txt")

	mock_task := NewSciTask("echo_foo_task", "", map[string]*FileTarget{"foo": NewFileTarget("foo.txt")}, nil, nil, nil, "")

	if p.PathFormatters["bar"](mock_task) != "foo.txt.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.txt.bar.txt"`)
	}
}

func TestSetPathReplace(t *testing.T) {
	p := NewFromShell("cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathReplace("foo", "bar", ".txt", ".bar.txt")

	mock_task := NewSciTask("echo_foo_task", "", map[string]*FileTarget{"foo": NewFileTarget("foo.txt")}, nil, nil, nil, "")

	if p.PathFormatters["bar"](mock_task) != "foo.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.bar.txt"`)
	}
}
