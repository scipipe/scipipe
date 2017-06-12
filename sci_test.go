package scipipe

import (
	"fmt"
	"os/exec"

	"github.com/stretchr/testify/assert"
	//"os"
	"os"
	t "testing"
	"time"
)

func initTestLogs() {
	//InitLogDebug()
	InitLogWarning()
}

func TestBasicRun(t *t.T) {
	initTestLogs()
	wf := NewWorkflow("TestBasicRunWf")

	t1 := NewProc("t1", "echo foo > {o:foo}")
	assert.IsType(t, t1.Out("foo"), NewFilePort())
	t1.PathFormatters["foo"] = func(t *SciTask) string {
		return "foo.txt"
	}
	wf.Add(t1)

	t2 := NewProc("t2", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	assert.IsType(t, t2.In("foo"), NewFilePort())
	assert.IsType(t, t2.Out("bar"), NewFilePort())
	t2.PathFormatters["bar"] = func(t *SciTask) string {
		return t.GetInPath("foo") + ".bar.txt"
	}
	wf.Add(t2)
	snk := NewSink("sink")

	t2.In("foo").Connect(t1.Out("foo"))
	snk.Connect(t2.Out("bar"))
	wf.SetSink(snk)

	assert.IsType(t, t2.In("foo"), NewFilePort())
	assert.IsType(t, t2.Out("bar"), NewFilePort())

	wf.Run()
	cleanFiles("foo.txt", "foo.txt.bar.txt")
}

func TestParameterCommand(t *t.T) {
	initTestLogs()
	wf := NewWorkflow("TestParameterCommandWf")

	cmb := NewCombinatoricsProcess("cmb")
	wf.Add(cmb)

	// An abc file printer
	abc := NewProc("abc", "echo {p:a} {p:b} {p:c} > {o:out}")
	abc.PathFormatters["out"] = func(task *SciTask) string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			task.Params["a"],
			task.Params["b"],
			task.Params["c"],
		)
	}
	wf.Add(abc)

	// A printer process
	prt := NewProc("prt", "cat {i:in} >> /tmp/log.txt; rm {i:in} {i:in}.audit.json")

	// Connection info
	abc.ParamPort("a").Connect(cmb.A)
	abc.ParamPort("b").Connect(cmb.B)
	abc.ParamPort("c").Connect(cmb.C)
	prt.In("in").Connect(abc.Out("out"))
	wf.SetDriver(prt)

	wf.Run()

	// Run tests
	_, err := os.Stat("/tmp/log.txt")
	assert.Nil(t, err)

	cleanFiles("/tmp/log.txt")
}

func TestProcessWithoutInputsOutputs(t *t.T) {
	initTestLogs()
	Debug.Println("Starting test TestProcessWithoutInputsOutputs")

	f := "/tmp/hej.txt"
	tsk := NewProc("tsk", "echo hej > "+f)
	tsk.Run()
	_, err := os.Stat(f)
	assert.Nil(t, err, fmt.Sprintf("File is missing: %s", f))
	cleanFiles(f)
}

func TestDontOverWriteExistingOutputs(t *t.T) {
	InitLogError()
	Debug.Println("Starting test TestDontOverWriteExistingOutputs")
	wf := NewWorkflow("TestDontOverWriteExistingOutputsWf1")

	f := "/tmp/hej.txt"

	// Assert file does not exist before running
	_, e1 := os.Stat(f)
	assert.NotNil(t, e1)

	// Run pipeline a first time
	tsk := NewProc("tsk", "echo hej > {o:hej1}")
	tsk.PathFormatters["hej1"] = func(task *SciTask) string { return f }
	wf.Add(tsk)

	prt := NewProc("prt", "echo {i:in1} Done!")
	prt.In("in1").Connect(tsk.Out("hej1"))
	wf.SetDriver(prt)

	wf.Run()

	// Assert file DO exist after running
	fiBef, e2 := os.Stat(f)
	assert.Nil(t, e2)

	// Get modified time before
	mtBef := fiBef.ModTime()

	// Make sure some time has passed before the second write
	time.Sleep(1 * time.Millisecond)

	Debug.Println("Try running the same workflow again ...")
	wf = NewWorkflow("TestDontOverWriteExistingOutputsWf2")

	// Run again with different output
	tsk = NewProc("tsk", "echo hej > {o:hej2}")
	tsk.PathFormatters["hej2"] = func(task *SciTask) string { return f }
	wf.Add(tsk)

	prt = NewProc("prt", "echo {i:in2} Done!")
	prt.In("in2").Connect(tsk.Out("hej2"))
	wf.SetDriver(prt)

	wf.Run()

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
	//InitLogDebug()

	fnames := []string{}
	for i := 1; i <= 10; i++ {
		fnames = append(fnames, fmt.Sprintf("/tmp/f%d.txt", i))
	}

	ig := NewIPGen("ipgen", fnames...)

	fc := NewProc("fc", "echo {i:in} > {o:out}")
	fc.SetPathExtend("in", "out", "")
	fc.In("in").Connect(ig.Out)

	sl := NewProc("sl", "cat {i:in} > {o:out}")
	sl.SetPathExtend("in", "out", ".copy.txt")
	sl.In("in").Connect(fc.Out("out"))

	//sl.Out("out").Chan = make(chan *InformationPacket, BUFSIZE)
	assert.NotNil(t, sl.Out)

	Debug.Println("TestSendsOrderedOutputs: Starting go-routines ...")

	go ig.Run()
	go fc.Run()
	go sl.Run()

	Debug.Println("TestSendsOrderedOutputs: Starting main loop ...")

	var expFname string
	i := 1

	tempPort := NewFilePort()
	Connect(tempPort, sl.Out("out"))

	for ft := range tempPort.Chan {
		Debug.Printf("TestSendsOrderedOutputs: Looping over item %d ...\n", i)
		expFname = fmt.Sprintf("/tmp/f%d.txt.copy.txt", i)
		assert.EqualValues(t, expFname, ft.GetPath())
		Debug.Printf("TestSendsOrderedOutputs: Looping over item %d Done.\n", i)
		i++
	}

	Debug.Println("TestSendsOrderedOutputs: Done with loop ...")

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
	wf := NewWorkflow("TestStreamingWf")

	// Init processes
	ls := NewProc("ls", "ls -l / > {os:lsl}")
	ls.PathFormatters["lsl"] = func(task *SciTask) string {
		return "/tmp/lsl.txt"
	}
	wf.Add(ls)

	grp := NewProc("grp", "grep etc {i:in} > {o:grepped}")
	grp.PathFormatters["grepped"] = func(task *SciTask) string {
		return task.GetInPath("in") + ".grepped.txt"
	}
	wf.Add(grp)

	snk := NewSink("sink")
	wf.SetSink(snk)

	// Connect
	grp.In("in").Connect(ls.Out("lsl"))
	snk.Connect(grp.Out("grepped"))

	// Run
	wf.Run()

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

func TestSubStreamReduceInPlaceHolder(t *t.T) {

	exec.Command("bash", "-c", "echo 1 > /tmp/file1.txt").CombinedOutput()
	exec.Command("bash", "-c", "echo 2 > /tmp/file2.txt").CombinedOutput()
	exec.Command("bash", "-c", "echo 3 > /tmp/file3.txt").CombinedOutput()

	wf := NewWorkflow("TestSubStreamReduceInPlaceHolderWf")

	// Create some input files

	ipg := NewIPGen("ipg", "/tmp/file1.txt", "/tmp/file2.txt", "/tmp/file3.txt")
	wf.Add(ipg)

	sts := NewStreamToSubStream()
	sts.In.Connect(ipg.Out)
	wf.Add(sts)

	cat := wf.NewProc("concatenate", "cat {i:infiles:r: } > {o:merged}")
	cat.SetPathStatic("merged", "/tmp/substream_merged.txt")
	cat.In("infiles").Connect(sts.OutSubStream)

	snk := NewSink("sink")
	snk.Connect(cat.Out("merged"))
	wf.SetSink(snk)

	wf.Run()

	_, err1 := os.Stat("/tmp/file1.txt")
	assert.Nil(t, err1, "File missing!")

	_, err2 := os.Stat("/tmp/file2.txt")
	assert.Nil(t, err2, "File missing!")

	_, err3 := os.Stat("/tmp/file3.txt")
	assert.Nil(t, err3, "File missing!")

	_, err4 := os.Stat("/tmp/substream_merged.txt")
	assert.Nil(t, err4, "File missing!")

	cleanFiles("/tmp/file1.txt", "/tmp/file2.txt", "/tmp/file3.txt", "/tmp/substream_merged.txt")
}

func TestMultipleLastProcs(t *t.T) {
	InitLogWarning()

	wf := NewWorkflow("TestMultipleLastProcs_WF")
	strs := []string{"hey", "how", "hoo"}

	for _, str := range strs {
		writeStr := wf.NewProc("writestr_"+str, "echo "+str+" > {o:out}")
		writeStr.SetPathStatic("out", "/tmp/"+str+".txt")

		catStr := wf.NewProc("catstr_"+str, "cat {i:in} > {o:out}")
		catStr.SetPathExtend("in", "out", ".cat.txt")

		catStr.In("in").Connect(writeStr.Out("out"))

		wf.ConnectLast(catStr.Out("out"))
	}

	wf.Run()

	for _, str := range strs {
		path := "/tmp/" + str + ".txt"
		_, err := os.Stat(path)
		assert.Nil(t, err, "File missing: "+path)
	}

	for _, str := range strs {
		path := "/tmp/" + str + ".txt.cat.txt"
		_, err := os.Stat(path)
		assert.Nil(t, err, "File missing: "+path)
	}

	cleanFiles("/tmp/hey.txt", "/tmp/how.txt", "/tmp/hoo.txt")
	cleanFiles("/tmp/hey.txt.cat.txt", "/tmp/how.txt.cat.txt", "/tmp/hoo.txt.cat.txt")
}

// Helper processes
type CombinatoricsProcess struct {
	Process
	name string
	A    *ParamPort
	B    *ParamPort
	C    *ParamPort
}

func NewCombinatoricsProcess(name string) *CombinatoricsProcess {
	return &CombinatoricsProcess{
		A:    NewParamPort(),
		B:    NewParamPort(),
		C:    NewParamPort(),
		name: name,
	}
}

func (proc *CombinatoricsProcess) Run() {
	defer proc.A.Close()
	defer proc.B.Close()
	defer proc.C.Close()

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				proc.A.Send(a)
				proc.B.Send(b)
				proc.C.Send(c)
			}
		}
	}
}

func (proc *CombinatoricsProcess) Name() string {
	return proc.name
}

func (proc *CombinatoricsProcess) IsConnected() bool { return true }

// StreamToSubstream helper process

type StreamToSubStream struct {
	In           *FilePort
	OutSubStream *FilePort
}

// Instantiate a new StreamToSubStream
func NewStreamToSubStream() *StreamToSubStream {
	return &StreamToSubStream{
		In:           NewFilePort(),
		OutSubStream: NewFilePort(),
	}
}

// Run the StreamToSubStream
func (proc *StreamToSubStream) Run() {
	defer proc.OutSubStream.Close()

	subStreamIP := NewInformationPacket("")
	subStreamIP.SubStream = proc.In

	proc.OutSubStream.Send(subStreamIP)
}

func (proc *StreamToSubStream) Name() string {
	return "StreamToSubstream"
}

func (proc *StreamToSubStream) IsConnected() bool {
	return proc.In.IsConnected() && proc.OutSubStream.IsConnected()
}

// Helper functions
func cleanFiles(fileNames ...string) {
	Debug.Println("Starting to remove files:", fileNames)
	for _, fileName := range fileNames {
		if _, err := os.Stat(fileName); err == nil {
			os.Remove(fileName)
			Debug.Println("Successfully removed file", fileName)
			// Remove any accompanying audit.json files ....
			os.Remove(fileName + ".audit.json")
			Debug.Println("Successfully removed audit.json file", fileName)
		}
	}
}
