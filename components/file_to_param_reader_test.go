package components

import (
	"os"
	"testing"

	"github.com/scipipe/scipipe"
)

func TestFileToParamReader(tt *testing.T) {
	// Create file to read
	filePath := "/tmp/filereader_testfile.txt"
	f, err := os.Create("/tmp/filereader_testfile.txt")
	if err != nil {
		tt.Fatalf("Could not create file: %s", filePath)
	}
	for _, s := range []string{"abc", "abc", "abc"} {
		f.WriteString(s + "\n")
	}
	f.Close()

	// Run test workflow and make sure that the parameter read from the file is
	// always "abc"
	wf := scipipe.NewWorkflow("wf", 4)
	rd := NewFileToParamReader(wf, "reader", filePath)
	checker := wf.NewProc("checker", "# {p:testparam}")
	checker.InParam("testparam").From(rd.OutLine())
	checker.CustomExecute = func(t *scipipe.Task) {
		expected := "abc"
		actual := t.Param("testparam")
		if actual != expected {
			tt.Errorf("Parameter was wrong. Actual: %s. Expected: %s", actual, expected)
		}
	}
	wf.Run()

	// Clean up test file
	err = os.Remove(filePath)
	if err != nil {
		tt.Fatalf("Could not remove file: %s", filePath)
	}
}
