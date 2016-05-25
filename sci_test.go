package scipipe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	//"os"
	"os"
	t "testing"
	"time"
)

func initTestLogs() {
	InitLogWarning()
}

func TestBasicRun(t *t.T) {
	initTestLogs()

	t1 := Shell("t1", "echo foo > {o:foo}")
	assert.IsType(t, t1.OutPorts["foo"], make(chan *FileTarget))
	t1.PathFormatters["foo"] = func(t *SciTask) string {
		return "foo.txt"
	}

	t2 := Shell("t2", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	assert.IsType(t, t2.InPorts["foo"], make(chan *FileTarget))
	assert.IsType(t, t2.OutPorts["bar"], make(chan *FileTarget))
	t2.PathFormatters["bar"] = func(t *SciTask) string {
		return t.GetInPath("foo") + ".bar.txt"
	}
	snk := NewSink()

	t2.InPorts["foo"] = t1.OutPorts["foo"]
	snk.In = t2.OutPorts["bar"]

	assert.IsType(t, t2.InPorts["foo"], make(chan *FileTarget))
	assert.IsType(t, t2.OutPorts["bar"], make(chan *FileTarget))

	pl := NewPipelineRunner()
	pl.AddProcesses(t1, t2, snk)
	pl.Run()

	cleanFiles("foo.txt", "foo.txt.bar.txt")
}

func TestParameterCommand(t *t.T) {
	initTestLogs()

	cmb := NewCombinatoricsProcess()

	// An abc file printer
	abc := Shell("abc", "echo {p:a} {p:b} {p:c} > {o:out}")
	abc.PathFormatters["out"] = func(task *SciTask) string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			task.Params["a"],
			task.Params["b"],
			task.Params["c"],
		)
	}

	// A printer process
	prt := Shell("prt", "cat {i:in} >> /tmp/log.txt; rm {i:in}")

	// Connection info
	abc.ParamPorts["a"] = cmb.A
	abc.ParamPorts["b"] = cmb.B
	abc.ParamPorts["c"] = cmb.C
	prt.InPorts["in"] = abc.OutPorts["out"]

	pl := NewPipelineRunner()
	pl.AddProcesses(cmb, abc, prt)
	pl.Run()

	// Run tests
	_, err := os.Stat("/tmp/log.txt")
	assert.Nil(t, err)

	cleanFiles("/tmp/log.txt")
}

func TestProcessWithoutInputsOutputs(t *t.T) {
	initTestLogs()
	Debug.Println("Starting test TestProcessWithoutInputsOutputs")

	f := "/tmp/hej.txt"
	tsk := Shell("tsk", "echo hej > "+f)
	tsk.Run()
	_, err := os.Stat(f)
	assert.Nil(t, err, fmt.Sprintf("File is missing: %s", f))
	cleanFiles(f)
}

func TestDontOverWriteExistingOutputs(t *t.T) {
	InitLogError()
	Debug.Println("Starting test TestDontOverWriteExistingOutputs")

	f := "/tmp/hej.txt"

	// Assert file does not exist before running
	_, e1 := os.Stat(f)
	assert.NotNil(t, e1)

	// Run pipeline a first time
	tsk := Shell("tsk", "echo hej > {o:hej}")
	tsk.PathFormatters["hej"] = func(task *SciTask) string { return f }
	prt := Shell("prt", "echo {i:in} Done!")
	prt.InPorts["in"] = tsk.OutPorts["hej"]
	pl := NewPipelineRunner()
	pl.AddProcesses(tsk, prt)
	pl.Run()

	// Assert file DO exist after running
	fiBef, e2 := os.Stat(f)
	assert.Nil(t, e2)

	// Get modified time before
	mtBef := fiBef.ModTime()

	// Make sure some time has passed before the second write
	time.Sleep(1 * time.Millisecond)

	Debug.Println("Try running the same workflow again ...")
	// Run again with different output
	tsk = Shell("tsk", "echo hej > {o:hej}")
	tsk.PathFormatters["hej"] = func(task *SciTask) string { return f }
	prt.InPorts["in"] = tsk.OutPorts["hej"]
	pl = NewPipelineRunner()
	pl.AddProcesses(tsk, prt)
	pl.Run()

	// Assert exists
	fiAft, e3 := os.Stat(f)
	assert.Nil(t, e3)

	// Get modified time AFTER second run
	mtAft := fiAft.ModTime()

	// Assert file is not modified!
	assert.EqualValues(t, mtBef, mtAft)

	cleanFiles(f)
}

// Make sure that outputs are returned in order, even though they are
// spawned to work in parallel.
func TestSendsOrderedOutputs(t *t.T) {
	initTestLogs()

	fnames := []string{}
	for i := 1; i <= 10; i++ {
		fnames = append(fnames, fmt.Sprintf("/tmp/f%d.txt", i))
	}

	fq := NewFileQueue(fnames...)

	fc := Shell("fc", "echo {i:in} > {o:out}")
	sl := Shell("sl", "cat {i:in} > {o:out}")

	fc.PathFormatters["out"] = func(task *SciTask) string { return task.GetInPath("in") }
	sl.PathFormatters["out"] = func(task *SciTask) string { return task.GetInPath("in") + ".copy.txt" }

	go fq.Run()
	go fc.Run()
	go sl.Run()

	fc.InPorts["in"] = fq.Out
	sl.InPorts["in"] = fc.OutPorts["out"]

	assert.NotEmpty(t, sl.OutPorts)

	var expFname string
	i := 1
	for ft := range sl.OutPorts["out"] {
		expFname = fmt.Sprintf("/tmp/f%d.txt.copy.txt", i)
		assert.EqualValues(t, expFname, ft.GetPath())
		i++
	}
	expFnames := []string{}
	for i := 1; i <= 10; i++ {
		expFnames = append(expFnames, fmt.Sprintf("/tmp/f%d.txt.copy.txt", i))
	}
	cleanFiles(fnames...)
	cleanFiles(expFnames...)
}

// Test that streaming works
func TestStreaming(t *t.T) {
	InitLogWarning()

	// Init processes
	ls := Shell("ls", "ls -l / > {os:lsl}")
	ls.PathFormatters["lsl"] = func(task *SciTask) string {
		return "/tmp/lsl.txt"
	}
	grp := Shell("grp", "grep etc {i:in} > {o:grepped}")
	grp.PathFormatters["grepped"] = func(task *SciTask) string {
		return task.GetInPath("in") + ".grepped.txt"
	}
	snk := NewSink()

	// Connect
	grp.InPorts["in"] = ls.OutPorts["lsl"]
	snk.In = grp.OutPorts["grepped"]

	// Run
	pl := NewPipelineRunner()
	pl.AddProcesses(ls, grp, snk)
	pl.Run()

	// Assert that a file exists
	_, err1 := os.Stat("/tmp/lsl.txt.fifo")
	assert.Nil(t, err1, "FIFO file does not exist, which it should!")

	// Assert otuput file exists
	_, err2 := os.Stat("/tmp/lsl.txt.grepped.txt")
	assert.Nil(t, err2, "File missing!")

	// Clean up
	cleanFiles("/tmp/lsl.txt", "/tmp/lsl.txt.grepped.txt")
	// cleanFiles("/tmp/lsl.txt.tmp")             // FIXME: Remove
	// cleanFiles("/tmp/lsl.txt.grepped.txt.tmp") // FIXME: Remove
	cleanFiles("/tmp/lsl.txt.fifo")
}

// Helper processes

type CombinatoricsProcess struct {
	Process
	A chan string
	B chan string
	C chan string
}

func NewCombinatoricsProcess() *CombinatoricsProcess {
	return &CombinatoricsProcess{
		A: make(chan string, BUFSIZE),
		B: make(chan string, BUFSIZE),
		C: make(chan string, BUFSIZE),
	}
}

func (proc *CombinatoricsProcess) Run() {
	defer close(proc.A)
	defer close(proc.B)
	defer close(proc.C)

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				proc.A <- a
				proc.B <- b
				proc.C <- c
			}
		}
	}
}

// Helper functions
func cleanFiles(fileNames ...string) {
	Debug.Println("Starting to remove files:", fileNames)
	for _, fileName := range fileNames {
		if _, err := os.Stat(fileName); err == nil {
			os.Remove(fileName)
			Debug.Println("Successfully removed file", fileName)
		}
	}
}
