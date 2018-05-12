package scipipe

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestSetWfName(t *testing.T) {
	initTestLogs()
	wf := NewWorkflow("TestWorkflow", 16)

	expectedWfName := "TestWorkflow"
	if wf.name != expectedWfName {
		t.Errorf("Workflow name is wrong, should be %s but is %s\n", wf.name, expectedWfName)
	}
}

func TestMaxConcurrentTasksCapacity(t *testing.T) {
	initTestLogs()
	wf := NewWorkflow("TestWorkflow", 16)

	if cap(wf.concurrentTasks) != 16 {
		t.Error("Wrong number of concurrent tasks")
	}
}

func TestAddProc(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("TestAddProcsWf", 16)

	proc1 := NewBogusProcess("bogusproc1")
	wf.AddProc(proc1)
	proc2 := NewBogusProcess("bogusproc2")
	wf.AddProc(proc2)

	if len(wf.procs) != 2 {
		t.Error("Wrong number of processes")
	}

	if !reflect.DeepEqual(reflect.TypeOf(wf.procs["bogusproc1"]), reflect.TypeOf(&BogusProcess{})) {
		t.Error("Bogusproc1 was not of the right type!")
	}
	if !reflect.DeepEqual(reflect.TypeOf(wf.procs["bogusproc2"]), reflect.TypeOf(&BogusProcess{})) {
		t.Error("Bogusproc2 was not of the right type!")
	}
}

func TestRunToProc(t *testing.T) {
	initTestLogs()

	wfa := getWorkflowForTestRunToProc("TestRunToProcWF_A")
	wfa.RunTo("mrg")

	if _, err := os.Stat("/tmp/foo.txt.bar.txt"); err != nil {
		t.Error("Merged file (/tmp/foo.txt.bar.txt) is not created, which it should")
	}

	if _, err := os.Stat("/tmp/foo.txt.bar.rpl.txt"); err == nil {
		t.Error("Replaced (merge) file (/tmp/foo.txt.bar.rpl.txt) exists, which it should not (yet)!")
	}

	time.Sleep(1 * time.Second)

	// We need to re-configure the workflow, since the connectivity will be
	// affected by the previous "RunTo" call
	wfb := getWorkflowForTestRunToProc("TestRunToProcWF_B")
	wfb.RunTo("rpl")

	if _, err := os.Stat("/tmp/foo.txt.bar.txt.rpl.txt"); err != nil {
		t.Errorf("Replaced (merge) file (/tmp/foo.txt.bar.rpl.txt) is not created, which it should (at this point): %v", err)
	}

	cleanFiles("/tmp/*.txt*")
}

func getWorkflowForTestRunToProc(wfName string) *Workflow {
	wf := NewWorkflow(wfName, 4)

	foo := wf.NewProc("foo", "echo foo > {o:out}")
	foo.SetPathStatic("out", "/tmp/foo.txt")

	bar := wf.NewProc("bar", "echo bar > {o:out}")
	bar.SetPathStatic("out", "/tmp/bar.txt")

	mrg := wf.NewProc("mrg", "cat {i:in1} {i:in2} > {o:mgd}")
	mrg.SetPathCustom("mgd", func(tk *Task) string {
		return tk.InPath("in1") + "." + filepath.Base(tk.InPath("in2"))
	})
	mrg.In("in1").Connect(foo.Out("out"))
	mrg.In("in2").Connect(bar.Out("out"))

	rpl := wf.NewProc("rpl", "cat {i:in} | sed 's/bar/baz/' > {o:out}")
	rpl.SetPathExtend("in", "out", ".rpl.txt")
	rpl.In("in").Connect(mrg.Out("mgd"))

	return wf
}

// --------------------------------
// Helper stuff
// --------------------------------

// A process with does just satisfy the Process interface, without doing any
// actual work.
type BogusProcess struct {
	BaseProcess
	name       string
	WasRun     bool
	WasRunLock sync.Mutex
}

func NewBogusProcess(name string) *BogusProcess {
	return &BogusProcess{WasRun: false, name: name}
}

func (p *BogusProcess) Run() {
	p.WasRunLock.Lock()
	p.WasRun = true
	p.WasRunLock.Unlock()
}

func (p *BogusProcess) Name() string {
	return p.name
}

func (p *BogusProcess) Connected() bool {
	return true
}
