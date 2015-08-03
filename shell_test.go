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
	InitLogWarn()
}

func TestBasicRun(t *t.T) {
	initTestLogs()

	t1 := Shell("echo foo > {o:foo}")
	assert.IsType(t, t1.OutPorts["foo"], make(chan *FileTarget))
	t1.OutPathFuncs["foo"] = func(t *ShellTask) string {
		return "foo.txt"
	}

	t2 := Shell("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	assert.IsType(t, t2.InPorts["foo"], make(chan *FileTarget))
	assert.IsType(t, t2.OutPorts["bar"], make(chan *FileTarget))
	t2.OutPathFuncs["bar"] = func(t *ShellTask) string {
		return t.GetInPath("foo") + ".bar.txt"
	}
	snk := NewSink()

	t2.InPorts["foo"] = t1.OutPorts["foo"]
	snk.In = t2.OutPorts["bar"]

	assert.IsType(t, t2.InPorts["foo"], make(chan *FileTarget))
	assert.IsType(t, t2.OutPorts["bar"], make(chan *FileTarget))

	pl := NewPipeline()
	pl.AddProcs(t1, t2, snk)
	pl.Run()

	cleanFiles("foo.txt", "foo.txt.bar.txt")
}

func TestParameterCommand(t *t.T) {
	initTestLogs()

	cmb := NewCombinatoricsProcess()

	// An abc file printer
	abc := Sh("echo {p:a} {p:b} {p:c} > {o:out}")
	abc.OutPathFuncs["out"] = func(task *ShellTask) string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			task.Params["a"],
			task.Params["b"],
			task.Params["c"],
		)
	}

	// A printer process
	prt := Sh("cat {i:in} >> /tmp/log.txt; rm {i:in}")

	// Connection info
	abc.ParamPorts["a"] = cmb.A
	abc.ParamPorts["b"] = cmb.B
	abc.ParamPorts["c"] = cmb.C
	prt.InPorts["in"] = abc.OutPorts["out"]

	pl := NewPipeline()
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
	tsk := Sh("echo hej > " + f)
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
	tsk := Sh("echo hej > {o:hej}")
	tsk.OutPathFuncs["hej"] = func(task *ShellTask) string { return f }
	prt := Sh("echo {i:in} Done!")
	prt.InPorts["in"] = tsk.OutPorts["hej"]
	pl := NewPipeline()
	pl.AddProcs(tsk, prt)
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
	tsk = Sh("echo hej > {o:hej}")
	tsk.OutPathFuncs["hej"] = func(task *ShellTask) string { return f }
	prt.InPorts["in"] = tsk.OutPorts["hej"]
	pl = NewPipeline()
	pl.AddProcs(tsk, prt)
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

	fc := Sh("echo {i:in} > {o:out}")
	sl := Sh("cat {i:in} > {o:out}")

	fc.OutPathFuncs["out"] = func(task *ShellTask) string { return task.GetInPath("in") }
	sl.OutPathFuncs["out"] = func(task *ShellTask) string { return task.GetInPath("in") + ".copy.txt" }

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
	InitLogWarn()

	// Init processes
	ls := Shell("ls -l / > {os:lsl}")
	ls.OutPathFuncs["lsl"] = func(task *ShellTask) string {
		return "/tmp/lsl.txt"
	}
	grp := Shell("grep etc {i:in} > {o:grepped}")
	grp.OutPathFuncs["grepped"] = func(task *ShellTask) string {
		return task.GetInPath("in") + ".grepped.txt"
	}
	snk := NewSink()

	// Connect
	grp.InPorts["in"] = ls.OutPorts["lsl"]
	snk.In = grp.OutPorts["grepped"]

	// Run
	pl := NewPipeline()
	pl.AddProcs(ls, grp, snk)
	pl.Run()

	// Assert no fifo file is left behind
	_, err1 := os.Stat("/tmp/lsl.txt.fifo")
	assert.NotNil(t, err1, "FIFO file exists, which should not!")

	// Assert otuput file exists
	_, err2 := os.Stat("/tmp/lsl.txt.grepped.txt")
	assert.Nil(t, err2, "File missing!")

	// Clean up
	cleanFiles("/tmp/lsl.txt", "/tmp/lsl.txt.grepped.txt")
	// cleanFiles("/tmp/lsl.txt.tmp")             // FIXME: Remove
	// cleanFiles("/tmp/lsl.txt.grepped.txt.tmp") // FIXME: Remove
	// cleanFiles("/tmp/lsl.txt.fifo")            // FIXME: Remove
}

// Helper processes

type CombinatoricsProcess struct {
	BaseProcess
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

	for _, a := range SS("a1", "a2", "a3") {
		for _, b := range SS("b1", "b2", "b3") {
			for _, c := range SS("c1", "c2", "c3") {
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
