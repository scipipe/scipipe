package scipipe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

func TestProcsSorted(t *testing.T) {
	wf := NewWorkflow("testwf", 4)
	wf.NewProc("p1", "#p1")
	wf.NewProc("p2", "#p2")
	wf.NewProc("p3", "#p3")
	wf.NewProc("p4", "#p4")

	actualNames := []string{}
	for _, p := range wf.ProcsSorted() {
		actualNames = append(actualNames, p.Name())
	}

	expectedNames := []string{"p1", "p2", "p3", "p4"}

	for i := range actualNames {
		if actualNames[i] != expectedNames[i] {
			t.Errorf("Presumedly sorted array wasnt as expected.\nEXPECTED:%v\nACTUAL:\n%v\n", expectedNames, actualNames)
		}
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

func TestDotGraph(t *testing.T) {
	initTestLogs()
	wf := NewWorkflow("testwf", 4)

	p1 := wf.NewProc("p1", "echo p1 > {o:out}")
	p1.SetOut("out", "/tmp/p1.txt")

	p2 := wf.NewProc("p2", "cat {i:in} > {o:out}")
	p2.SetOut("out", "{i:in}.p2.txt")
	p2.In("in").From(p1.Out("out"))

	expectedTailLabels := `digraph "testwf" {
  graph [fontname="Arial",fontsize=13,color="#384A52",fontcolor="#384A52"];
  node  [fontname="Arial",fontsize=11,color="#384A52",fontcolor="#384A52",fillcolor="#EFF2F5",shape=box,style=filled];
  edge  [fontname="Arial",fontsize=9, color="#384A52",fontcolor="#384A52"];
  "p1" [shape=box];
  "p2" [shape=box];
  "p1" -> "p2" [taillabel="out", headlabel="in"];
}
`
	actualTailLabels := wf.DotGraph()
	if actualTailLabels != expectedTailLabels {
		t.Errorf("Dot graph with tail label is not as expected!\nEXPECTED:\n%s\nACTUAL:\n%s\n", expectedTailLabels, actualTailLabels)
	}

	wf.PlotConf.EdgeLabels = false
	expected := `digraph "testwf" {
  graph [fontname="Arial",fontsize=13,color="#384A52",fontcolor="#384A52"];
  node  [fontname="Arial",fontsize=11,color="#384A52",fontcolor="#384A52",fillcolor="#EFF2F5",shape=box,style=filled];
  edge  [fontname="Arial",fontsize=9, color="#384A52",fontcolor="#384A52"];
  "p1" [shape=box];
  "p2" [shape=box];
  "p1" -> "p2";
}
`
	actual := wf.DotGraph()
	if expected != actual {
		t.Errorf("Dot graph is not as expected!\nEXPECTED:\n%s\nACTUAL:\n%s\n", expected, actual)
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

	cleanFilePatterns("/tmp/foo.txt*")
}

func getWorkflowForTestRunToProc(wfName string) *Workflow {
	wf := NewWorkflow(wfName, 4)

	foo := wf.NewProc("foo", "echo foo > {o:out}")
	foo.SetOut("out", "/tmp/foo.txt")

	bar := wf.NewProc("bar", "echo bar > {o:out}")
	bar.SetOut("out", "/tmp/bar.txt")

	mrg := wf.NewProc("mrg", "cat {i:in1} {i:in2} > {o:mgd}")
	mrg.SetOutFunc("mgd", func(tk *Task) string {
		return tk.InPath("in1") + "." + filepath.Base(tk.InPath("in2"))
	})
	mrg.In("in1").From(foo.Out("out"))
	mrg.In("in2").From(bar.Out("out"))

	rpl := wf.NewProc("rpl", "cat {i:in} | sed 's/bar/baz/' > {o:out}")
	rpl.SetOut("out", "{i:in}.rpl.txt")
	rpl.In("in").From(mrg.Out("mgd"))

	return wf
}

func TestBasicRun(t *testing.T) {
	initTestLogs()
	wf := NewWorkflow("TestBasicRunWf", 16)

	p1 := wf.NewProc("p1", "echo foo > {o:foo}")
	assertIsType(t, p1.Out("foo"), NewOutPort("foo"))
	p1.SetOutFunc("foo", func(t *Task) string { return "foo.txt" })

	p2 := wf.NewProc("p2", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	assertIsType(t, p2.In("foo"), NewInPort("foo"))
	assertIsType(t, p2.Out("bar"), NewOutPort("bar"))
	p2.SetOut("bar", "{i:foo}.bar.txt")

	p2.In("foo").From(p1.Out("foo"))

	assertIsType(t, p2.In("foo"), NewInPort("foo"))
	assertIsType(t, p2.Out("bar"), NewOutPort("bar"))

	wf.Run()
	cleanFiles("foo.txt", "foo.txt.bar.txt")
}

func TestWithoutPathFormatter(t *testing.T) {
	initTestLogs()
	wf := NewWorkflow("TestWithoutPathFormatter", 4)
	foo := wf.NewProc("foo", "echo foo > {o:txt}")
	f2b := wf.NewProc("f2b", "cat {i:in} | sed 's/foo/bar/g' > {o:txt}")
	f2b.In("in").From(foo.Out("txt"))
	wf.Run()
	fps := []string{"foo.txt", "foo.txt.audit.json", "foo.txt.f2b.txt", "foo.txt.f2b.txt.audit.json"}
	for _, fp := range fps {
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("File does not exist: %s\n", fp)
		}
	}
	cleanFiles(fps...)
}

func TestParameterCommand(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("TestParameterCommandWf", 4)

	cmb := NewCombinatoricsProcess("cmb")
	wf.AddProc(cmb)

	// An abc file printer
	abc := wf.NewProc("abc", "echo {p:a} {p:b} {p:c} > {o:out}")
	abc.SetOut("out", "/tmp/abc_{p:a}_{p:b}_{p:c}.txt")
	abc.InParam("a").From(cmb.A)
	abc.InParam("b").From(cmb.B)
	abc.InParam("c").From(cmb.C)

	// A printer process
	prt := wf.NewProc("prt", "cat {i:in} >> /tmp/log.txt")
	prt.In("in").From(abc.Out("out"))

	wf.Run()

	// Run tests
	filePath := "/tmp/log.txt"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File does not exist: %s\n", filePath)
	}

	cleanFiles(filePath)
	cleanFilePatterns("/tmp/abc_*_*_*.txt*")
}

func TestDontOverWriteExistingOutputs(t *testing.T) {
	initTestLogs()
	Debug.Println("Starting test TestDontOverWriteExistingOutputs")

	f := "/tmp/hej.txt"

	// Assert file does not exist before running
	_, e1 := os.Stat(f)
	assertNotNil(t, e1)

	// Run pipeline a first time
	wf1 := NewWorkflow("TestDontOverWriteExistingOutputsWf1", 16)
	tsk := wf1.NewProc("tsk", "echo hej > {o:hej1}")
	tsk.SetOut("hej1", f)
	prt := wf1.NewProc("prt", "cat {i:in1} > {o:done}")
	prt.SetOut("done", "{i:in1}.done.txt")
	prt.In("in1").From(tsk.Out("hej1"))
	wf1.Run()

	// Assert file DO exist after running
	fiBef, e2 := os.Stat(f)
	assertNil(t, e2)

	time.Sleep(1 * time.Millisecond)
	// Get modified time before
	mtBef := fiBef.ModTime()

	// Run again with different output
	wf2 := NewWorkflow("TestDontOverWriteExistingOutputsWf2", 16)
	tsk = wf2.NewProc("tsk", "echo hej > {o:hej2}")
	tsk.SetOutFunc("hej2", func(task *Task) string { return f })
	prt = wf2.NewProc("prt", "cat {i:in1} > {o:done}")
	prt.SetOut("done", "{i:in1}.done.txt")
	prt.In("in1").From(tsk.Out("hej2"))
	wf2.Run()

	// Assert exists
	fiAft, e3 := os.Stat(f)
	assertNil(t, e3)

	// Get modified time AFTER second run
	mtAft := fiAft.ModTime()

	// Assert file is not modified!
	assertEqualValues(t, mtBef, mtAft)

	cleanFiles(f, f+".done.txt")
}

// Make sure that outputs are returned in order, even though they are
// spawned to work in parallel.
func TestSendOrderedOutputs(t *testing.T) {
	initTestLogs()

	fnames := []string{}
	for i := 1; i <= 10; i++ {
		fnames = append(fnames, fmt.Sprintf("/tmp/f%d.txt", i))
	}

	wf := NewWorkflow("test_wf", 16)
	ig := NewFileSource(wf, "ipgen", fnames...)

	fc := wf.NewProc("fc", "echo {i:in} > {o:out}")
	fc.SetOut("out", "{i:in}")
	fc.In("in").From(ig.Out())

	sl := wf.NewProc("sl", "cat {i:in} > {o:out}")
	sl.SetOut("out", "{i:in}.copy.txt")
	sl.In("in").From(fc.Out("out"))

	assertNotNil(t, sl.Out)

	var expFname string
	i := 1

	tempPort := NewInPort("temp")
	tempPort.process = NewBogusProcess("bogus_process")
	tempPort.From(sl.Out("out"))

	// Should not start go-routines before connection stuff is done
	go ig.Run()
	go fc.Run()
	go sl.Run()

	for ft := range tempPort.Chan {
		Debug.Printf("TestSendOrderedOutputs: Looping over item %d ...\n", i)
		expFname = fmt.Sprintf("/tmp/f%d.txt.copy.txt", i)
		assertEqualValues(t, expFname, ft.Path())
		Debug.Printf("TestSendOrderedOutputs: Looping over item %d Done.\n", i)
		i++
	}

	Debug.Println("TestSendOrderedOutputs: Done with loop ...")

	expFnames := []string{}
	for i := 1; i <= 10; i++ {
		expFnames = append(expFnames, fmt.Sprintf("/tmp/f%d.txt.copy.txt", i))
	}
	cleanFiles(fnames...)
	cleanFiles(expFnames...)
}

// Test that streaming works
func TestStreaming(t *testing.T) {
	initTestLogs()

	// Set up and run workflow
	wf := NewWorkflow("TestStreamingWf", 16)
	ls := wf.NewProc("ls", "ls -l / > {os:lsl}")
	ls.SetOut("lsl", "/tmp/lsl.txt")
	grp := wf.NewProc("grp", "grep etc {i:in} > {o:grepped}")
	grp.SetOut("grepped", "{i:in}.grepped.txt")
	grp.In("in").From(ls.Out("lsl"))
	wf.Run()

	// Assert otuput file exists
	_, err2 := os.Stat("/tmp/lsl.txt.grepped.txt")
	assertNil(t, err2, "File missing!")

	// Clean up
	cleanFiles("/tmp/lsl.txt", "/tmp/lsl.txt.grepped.txt")
}

func TestSubStreamJoinInPlaceHolder(t *testing.T) {
	initTestLogs()

	exec.Command("bash", "-c", "echo 1 > /tmp/file1.txt").CombinedOutput()
	exec.Command("bash", "-c", "echo 2 > /tmp/file2.txt").CombinedOutput()
	exec.Command("bash", "-c", "echo 3 > /tmp/file3.txt").CombinedOutput()

	wf := NewWorkflow("TestSubStreamJoinInPlaceHolderWf", 16)

	// Create some input files

	ipg := NewFileSource(wf, "ipg", "/tmp/file1.txt", "/tmp/file2.txt", "/tmp/file3.txt")

	sts := NewStreamToSubStream(wf, "str_to_substr")
	sts.In().From(ipg.Out())

	cat := wf.NewProc("concatenate", "cat {i:infiles|join: } > {o:merged}")
	cat.SetOut("merged", "/tmp/substream_merged.txt")
	cat.In("infiles").From(sts.OutSubStream())

	wf.Run()

	_, err1 := os.Stat("/tmp/file1.txt")
	assertNil(t, err1, "File missing!")

	_, err2 := os.Stat("/tmp/file2.txt")
	assertNil(t, err2, "File missing!")

	_, err3 := os.Stat("/tmp/file3.txt")
	assertNil(t, err3, "File missing!")

	_, err4 := os.Stat("/tmp/substream_merged.txt")
	assertNil(t, err4, "File missing!")

	cleanFiles("/tmp/file1.txt", "/tmp/file2.txt", "/tmp/file3.txt", "/tmp/substream_merged.txt")
}

func TestMultipleLastProcs(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("TestMultipleLastProcs_WF", 16)
	strs := []string{"hey", "how", "hoo"}

	for _, str := range strs {
		writeStr := wf.NewProc("writestr_"+str, "echo "+str+" > {o:out}")
		writeStr.SetOut("out", "/tmp/"+str+".txt")

		catStr := wf.NewProc("catstr_"+str, "cat {i:in} > {o:out}")
		catStr.SetOut("out", "{i:in}.cat.txt")
		catStr.In("in").From(writeStr.Out("out"))
	}

	wf.Run()

	for _, str := range strs {
		path := "/tmp/" + str + ".txt"
		_, err := os.Stat(path)
		assertNil(t, err, "File missing: "+path)
	}

	for _, str := range strs {
		path := "/tmp/" + str + ".txt.cat.txt"
		_, err := os.Stat(path)
		assertNil(t, err, "File missing: "+path)
	}

	cleanFiles("/tmp/hey.txt", "/tmp/how.txt", "/tmp/hoo.txt")
	cleanFiles("/tmp/hey.txt.cat.txt", "/tmp/how.txt.cat.txt", "/tmp/hoo.txt.cat.txt")
}

func TestPassOnTags(t *testing.T) {
	wf := NewWorkflow("TestPassOnTags_WF", 4)

	hey := wf.NewProc("create_file", "echo hey > {o:heyfile}")
	hey.SetOut("heyfile", "/tmp/hey.txt")

	tag := NewMapToTags(wf, "add_tag", func(ip *FileIP) map[string]string {
		return map[string]string{"hey": "you"}
	})
	tag.In().From(hey.Out("heyfile"))

	you := wf.NewProc("add_you", "echo '$(cat {i:infile}) you' > {o:youfile}")
	you.SetOut("youfile", "{i:infile}.you.txt")
	you.In("infile").From(tag.Out())

	wf.Run()

	dat, err := ioutil.ReadFile("/tmp/hey.txt.you.txt.audit.json")
	Check(err)
	auditInfo := &AuditInfo{}
	err = json.Unmarshal(dat, auditInfo)
	Check(err)

	assertEqualValues(t, "you", auditInfo.Tags["hey"], "Audit info does not contain passed on tags")

	cleanFiles("/tmp/hey.txt", "/tmp/hey.txt.you.txt")
}

// TestReceiveBothIPsAndParams makes sure that channels in the process
// createTask process are not short-cut before all parameters and IPs are
// received, by running a workflow that receives both a stream of params, and
// one or more IPs
func TestReceiveBothIPsAndParams(t *testing.T) {
	wf := NewWorkflow("multiout", 4)

	echo := wf.NewProc("echo", "echo hej > {o:hej}")
	echo.SetOut("hej", "/tmp/ipsparams.hej.txt")

	params := NewParamSource(wf, "params", "tjo", "hej", "hopp")

	strs := []string{"foo", "bar", "baz"}
	for _, str := range strs {
		add := wf.NewProc("add_"+str, `echo {i:infile} {p:param} > {o:out}`)
		str := str
		add.SetOutFunc("out", func(t *Task) string {
			return t.InPath("infile") + "." + str + ".txt"
		})
		add.InParam("param").From(params.Out())
		add.In("infile").From(echo.Out("hej"))
	}

	wf.Run()

	_, err := os.Stat("/tmp/ipsparams.hej.txt")
	assertNil(t, err)

	files := []string{"/tmp/ipsparams.hej.txt"}
	for _, str := range strs {
		file := "/tmp/ipsparams.hej.txt." + str + ".txt"
		_, err := os.Stat(file)
		assertNil(t, err)
		files = append(files, file)
	}

	cleanFiles(files...)
}

func TestSanitizePathFragments(t *testing.T) {
	for input, expected := range map[string]string{
		"Base Complement": "base_complement",
	} {
		actual := sanitizePathFragment(input)
		if actual != expected {
			t.Errorf("Result was not as expected:\nEXPECTED:\n%s\nACTUAL:\n%s", expected, actual)
		}
	}
}

// --------------------------------------------------------------------------------
// CombinatoricsProcess helper process
// --------------------------------------------------------------------------------

type CombinatoricsProcess struct {
	name string
	A    *OutParamPort
	B    *OutParamPort
	C    *OutParamPort
}

func NewCombinatoricsProcess(name string) *CombinatoricsProcess {
	a := NewOutParamPort("a")
	b := NewOutParamPort("b")
	c := NewOutParamPort("c")
	p := &CombinatoricsProcess{
		A:    a,
		B:    b,
		C:    c,
		name: name,
	}
	a.process = p
	b.process = p
	c.process = p
	return p
}

func (p *CombinatoricsProcess) InPorts() map[string]*InPort {
	return map[string]*InPort{}
}
func (p *CombinatoricsProcess) OutPorts() map[string]*OutPort {
	return map[string]*OutPort{}
}
func (p *CombinatoricsProcess) InParamPorts() map[string]*InParamPort {
	return map[string]*InParamPort{}
}
func (p *CombinatoricsProcess) OutParamPorts() map[string]*OutParamPort {
	return map[string]*OutParamPort{
		p.A.Name(): p.A,
		p.B.Name(): p.B,
		p.C.Name(): p.C,
	}
}

func (p *CombinatoricsProcess) Run() {
	defer p.A.Close()
	defer p.B.Close()
	defer p.C.Close()

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				p.A.Send(a)
				p.B.Send(b)
				p.C.Send(c)
			}
		}
	}
}

func (p *CombinatoricsProcess) Name() string {
	return p.name
}

func (p *CombinatoricsProcess) Ready() bool { return true }

// --------------------------------------------------------------------------------
// StreamToSubstream helper process
// --------------------------------------------------------------------------------

type StreamToSubStream struct {
	BaseProcess
}

func NewStreamToSubStream(wf *Workflow, name string) *StreamToSubStream {
	p := &StreamToSubStream{
		BaseProcess: NewBaseProcess(wf, name),
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "substream")
	wf.AddProc(p)
	return p
}

func (p *StreamToSubStream) In() *InPort            { return p.InPort("in") }
func (p *StreamToSubStream) OutSubStream() *OutPort { return p.OutPort("substream") }

func (p *StreamToSubStream) Run() {
	defer p.OutSubStream().Close()

	subStreamIP := NewFileIP("")
	subStreamIP.SubStream = p.In()

	p.OutSubStream().Send(subStreamIP)
}

// --------------------------------------------------------------------------------
// MapToTag helper process
// --------------------------------------------------------------------------------

type MapToTags struct {
	BaseProcess
	mapFunc func(ip *FileIP) map[string]string
}

func NewMapToTags(wf *Workflow, name string, mapFunc func(ip *FileIP) map[string]string) *MapToTags {
	p := &MapToTags{
		BaseProcess: NewBaseProcess(wf, name),
		mapFunc:     mapFunc,
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

func (p *MapToTags) In() *InPort   { return p.InPort("in") }
func (p *MapToTags) Out() *OutPort { return p.OutPort("out") }

func (p *MapToTags) Run() {
	defer p.CloseAllOutPorts()
	for ip := range p.In().Chan {
		newTags := p.mapFunc(ip)
		ip.AddTags(newTags)
		ip.WriteAuditLogToFile()
		p.Out().Send(ip)
	}
}

// --------------------------------------------------------------------------------
// FileSource helper process
// --------------------------------------------------------------------------------

// FileSource is initiated with a set of file paths, which it will send as a
// stream of File IPs on its outport Out()
type FileSource struct {
	BaseProcess
	filePaths []string
}

// NewFileSource returns a new initialized FileSource process
func NewFileSource(wf *Workflow, name string, filePaths ...string) *FileSource {
	p := &FileSource{
		BaseProcess: NewBaseProcess(wf, name),
		filePaths:   filePaths,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port, on which file IPs based on the file paths the
// process was initialized with, will be retrieved.
func (p *FileSource) Out() *OutPort { return p.OutPort("out") }

// Run runs the FileSource process
func (p *FileSource) Run() {
	defer p.CloseAllOutPorts()
	for _, filePath := range p.filePaths {
		p.Out().Send(NewFileIP(filePath))
	}
}

// --------------------------------------------------------------------------------
// ParamSource helper process
// --------------------------------------------------------------------------------

// ParamSource will feed parameters on an out-port
type ParamSource struct {
	BaseProcess
	params []string
}

// NewParamSource returns a new ParamSource
func NewParamSource(wf *Workflow, name string, params ...string) *ParamSource {
	p := &ParamSource{
		BaseProcess: NewBaseProcess(wf, name),
		params:      params,
	}
	p.InitOutParamPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port, on which parameters the process was initialized
// with, will be retrieved.
func (p *ParamSource) Out() *OutParamPort { return p.OutParamPort("out") }

// Run runs the process
func (p *ParamSource) Run() {
	defer p.CloseAllOutPorts()
	for _, param := range p.params {
		p.Out().Send(param)
	}
}

// --------------------------------
// BogusProcess helper process
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

func (p *BogusProcess) Ready() bool {
	return true
}
