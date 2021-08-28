package components

import (
	"fmt"
	"os"
	"testing"

	"github.com/scipipe/scipipe"
)

var params = []string{"abc", "bcd", "cde"}

func TestFileToParamsReader(tt *testing.T) {
	os.MkdirAll(".tmp", 0744)
	// Create file to read
	filePath := ".tmp/filereader_testfile.txt"
	f, err := os.Create(".tmp/filereader_testfile.txt")
	if err != nil {
		tt.Fatalf("Could not create file: %s", filePath)
	}
	for _, s := range params {
		f.WriteString(s + "\n")
	}
	f.Close()

	// Run test workflow and make sure that the parameter read from the file is
	// always "abc"
	wf := scipipe.NewWorkflow("wf", 4)
	rd := NewFileToParamsReader(wf, "reader", filePath)
	checker := NewFileToParamsChecker(wf, "filetoparams_checker", tt)
	checker.InParams().From(rd.OutLine())

	wf.Run()

	// Clean up test file
	err = os.Remove(filePath)
	if err != nil {
		tt.Fatalf("Could not remove file: %s", filePath)
	}
}

type FileToParamsChecker struct {
	scipipe.BaseProcess
	*testing.T
}

func NewFileToParamsChecker(wf *scipipe.Workflow, pname string, t *testing.T) *FileToParamsChecker {
	p := &FileToParamsChecker{
		scipipe.NewBaseProcess(wf, pname),
		t,
	}
	p.InitInParamPort(p, "params")
	wf.AddProc(p)
	return p
}

func (p *FileToParamsChecker) InParams() *scipipe.InParamPort { return p.InParamPort("params") }

func (p *FileToParamsChecker) Run() {
	i := 0
	for param := range p.InParams().Chan {
		expected := params[i]
		actual := param
		if actual != expected {
			p.T.Errorf("actual parameter value (%s) was not as expected (%s) in FileToParamsReader", actual, expected)
		}
		i++
	}
}

func (p *FileToParamsChecker) Failf(msg string, parts ...interface{}) {
	p.Fail(fmt.Sprintf(msg, parts...))
}

func (p *FileToParamsChecker) Fail(msg interface{}) {
	scipipe.Failf("[Process:%s] %s", p.Name(), msg)
}
