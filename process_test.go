package scipipe

import (
	"testing"
)

func TestShell(t *testing.T) {
	p1 := Shell("echo", "echo {p:text}")
	if p1.ParamPorts["text"] == nil {
		t.Error(`p.ParamPorts["text"] = nil. want: not nil`)
	}

	p2 := Shell("cat", "cat {i:infile} > {o:outfile}")
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

func TestSetPathFormatStatic(t *testing.T) {
	p := Shell("echo_foo", "echo foo > {o:bar}")
	p.SetPathFormatStatic("bar", "bar.txt")

	mock_task := NewSciTask("echo_foo_task", "", nil, nil, nil, nil, "")

	if p.PathFormatters["bar"](mock_task) != "bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "bar.txt"`)
	}
}

func TestSetPathFormatExtend(t *testing.T) {
	p := Shell("cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathFormatExtend("foo", "bar", ".bar.txt")

	mock_task := NewSciTask("echo_foo_task", "", map[string]*FileTarget{"foo": NewFileTarget("foo.txt")}, nil, nil, nil, "")

	if p.PathFormatters["bar"](mock_task) != "foo.txt.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.txt.bar.txt"`)
	}
}

func TestSetPathFormatReplace(t *testing.T) {
	p := Shell("cat_foo", "cat {i:foo} > {o:bar}")
	p.SetPathFormatReplace("foo", "bar", ".txt", ".bar.txt")

	mock_task := NewSciTask("echo_foo_task", "", map[string]*FileTarget{"foo": NewFileTarget("foo.txt")}, nil, nil, nil, "")

	if p.PathFormatters["bar"](mock_task) != "foo.bar.txt" {
		t.Error(`p.PathFormatters["bar"]() != "foo.bar.txt"`)
	}
}
