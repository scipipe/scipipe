package scipipe

import (
	"errors"
	"testing"
)

func TestExecCmd_EchoFooBar(t *testing.T) {
	output := ExecCmd("echo foo bar")
	if output != "foo bar\n" {
		t.Errorf("output = %swant: foo bar\n", output)
	}
}

func TestCheck_Panics(t *testing.T) {
	// Recover the panic, and check that the recover "was needed" (r was not
	// nil)
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic as it should!")
		}
	}()
	err := errors.New("A test-error")
	Check(err)
}
