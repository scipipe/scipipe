package components

import (
	"testing"

	"github.com/scipipe/scipipe"
)

func TestCommandToParams(tt *testing.T) {
	// Run test workflow and make sure that the parameter read from the file is
	// always "abc"
	wf := scipipe.NewWorkflow("wf", 4)

	cmdToParams := NewCommandToParams(wf, "cmdtoparams", "echo foo; echo bar; echo baz;")

	checker := wf.NewProc("checker", "# {p:param}")
	checker.CustomExecute = func(t *scipipe.Task) {
		expected := []string{"foo", "bar", "baz"}
		actual := t.Param("param")
		if actual != expected[0] && actual != expected[1] && actual != expected[2] {
			tt.Errorf("Actual string (%s) was not one of the expected ones (%s, %s, %s)", actual, expected[0], expected[1], expected[2])
		}
	}
	checker.InParam("param").From(cmdToParams.OutParam())

	wf.Run()
}
