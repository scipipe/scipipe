// Package scipipe is a library for writing scientific workflows (sometimes
// also called "pipelines") of shell commands that depend on each other, in the
// Go programming languages. It was initially designed for problems in
// cheminformatics and bioinformatics, but should apply equally well to any
// domain involving complex pipelines of interdependent shell commands.
package scipipe

import (
	"os"
)

// ----------------------------------------------------------------------------
// Workflow
// ----------------------------------------------------------------------------

type Workflow struct {
	name   string
	procs  map[string]Process
	sink   *Sink
	driver Process
}

func NewWorkflow(name string) *Workflow {
	if !LogExists {
		InitLogInfo()
	}
	sink := NewSink(name + "_default_sink")
	return &Workflow{
		name:   name,
		procs:  map[string]Process{},
		sink:   sink,
		driver: sink,
	}
}

func (wf *Workflow) Add(proc Process) {
	if wf.procs[proc.Name()] != nil {
		Error.Fatalf(wf.name+" workflow: A process with name '%s' already exists in the workflow! Use a more unique name!\n", proc.Name())
	}
	wf.procs[proc.Name()] = proc
}

func (wf *Workflow) NewProc(procName string, commandPattern string) *SciProcess {
	proc := NewProc(procName, commandPattern)
	wf.Add(proc)
	return proc
}

func (wf *Workflow) AddProcs(procs ...Process) {
	for _, proc := range procs {
		wf.procs[proc.Name()] = proc
	}
}

func (wf *Workflow) Proc(procName string) Process {
	return wf.procs[procName]
}

func (wf *Workflow) Procs() map[string]Process {
	return wf.procs
}

func (wf *Workflow) Sink() *Sink {
	return wf.sink
}

func (wf *Workflow) SetSink(sink *Sink) {
	if wf.sink.IsConnected() {
		Error.Println("Trying to replace a sink which is already connected. Are you combining SetSink() with ConnectFinalOutPort()? That is not allowed!")
		os.Exit(1)
	}
	wf.sink = sink
	wf.driver = sink
}

func (wf *Workflow) Driver() Process {
	return wf.driver
}

func (wf *Workflow) SetDriver(driver Process) {
	wf.driver = driver
}

// ConnectLast connects the last (most downstream) out-ports in the workflow to
// an implicit sink process which will be used to drive the workflow. This can
// be used instead of manually creating a sink, connecting it, and setting it
// as the driver process of the workflow.
func (wf *Workflow) ConnectLast(outPort *FilePort) {
	wf.sink.Connect(outPort)
	// Make sure the sink is also the driver
	wf.driver = wf.sink
}

func (wf *Workflow) Run() {
	if len(wf.procs) == 0 {
		Error.Println(wf.name + ": The workflow is empty. Did you forget to add the processes to it?")
		os.Exit(1)
	}
	if wf.sink == nil {
		Error.Println(wf.name + ": sink is nil!")
		os.Exit(1)
	}
	for _, proc := range wf.procs {
		if !proc.IsConnected() {
			Error.Println(wf.name + ": Not everything connected. Workflow shutting down.")
			os.Exit(1)
		}
	}
	for pname, proc := range wf.procs {
		Debug.Printf(wf.name+": Starting process %s in new go-routine", pname)
		go proc.Run()
	}
	Debug.Printf(wf.name + ": Starting sink in main go-routine")
	wf.driver.Run()
}
