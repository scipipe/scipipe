// Package scipipe is a library for writing scientific workflows (sometimes
// also called "pipelines") of shell commands that depend on each other, in the
// Go programming languages. It was initially designed for problems in
// cheminformatics and bioinformatics, but should apply equally well to any
// domain involving complex pipelines of interdependent shell commands.
package scipipe

import (
	"sync"
)

// ----------------------------------------------------------------------------
// Workflow
// ----------------------------------------------------------------------------

// Workflow is the centerpiece of the functionality in SciPipe, and is a
// container for a pipeline of processes making up a workflow. It has various
// methods for coordination the execution of the pipeline as a whole, such as
// keeping track of the maxiumum number of concurrent tasks, as well as helper
// methods for creating new processes, that automatically gets plugged in to the
// workflow on creation
type Workflow struct {
	name              string
	procs             map[string]WorkflowProcess
	concurrentTasks   chan struct{}
	concurrentTasksMx sync.Mutex
	sink              *Sink
	driver            WorkflowProcess
}

// NewWorkflow returns a new Workflow
func NewWorkflow(name string, maxConcurrentTasks int) *Workflow {
	InitLogInfo()
	sink := NewSink(name + "_default_sink")
	return &Workflow{
		name:            name,
		procs:           map[string]WorkflowProcess{},
		concurrentTasks: make(chan struct{}, maxConcurrentTasks),
		sink:            sink,
		driver:          sink,
	}
}

// WorkflowProcess is an interface for processes to be handled by Workflow
type WorkflowProcess interface {
	Name() string
	IsConnected() bool
	Run()
}

// AddProc adds a Process to the workflow, to be run when the workflow runs
func (wf *Workflow) AddProc(proc WorkflowProcess) {
	if wf.procs[proc.Name()] != nil {
		Error.Fatalf(wf.name+" workflow: A process with name '%s' already exists in the workflow! Use a more unique name!\n", proc.Name())
	}
	wf.procs[proc.Name()] = proc
}

// AddProcs takes one or many Processes and adds them to the workflow, to be run
// when the workflow runs.
func (wf *Workflow) AddProcs(procs ...WorkflowProcess) {
	for _, proc := range procs {
		wf.AddProc(proc)
	}
}

// NewProc returns a new process based on a commandPattern (See the
// documentation for scipipe.NewProcess for more details about the pattern) and
// connects the process to the workflow
func (wf *Workflow) NewProc(procName string, commandPattern string) *Process {
	proc := NewProc(wf, procName, commandPattern)
	return proc
}

// Proc returns the process with name procName from the workflow
func (wf *Workflow) Proc(procName string) WorkflowProcess {
	return wf.procs[procName]
}

// Procs returns a map of all processes keyed by their names in the workflow
func (wf *Workflow) Procs() map[string]WorkflowProcess {
	return wf.procs
}

// Sink returns the sink process of the workflow
func (wf *Workflow) Sink() *Sink {
	return wf.sink
}

// SetSink sets the sink of the workflow to the provided sink process
func (wf *Workflow) SetSink(sink *Sink) {
	if wf.sink.IsConnected() {
		Error.Fatalln("Trying to replace a sink which is already connected. Are you combining SetSink() with ConnectFinalOutPort()? That is not allowed!")
	}
	wf.sink = sink
	wf.driver = sink
}

// Driver returns the driver process of the workflow
func (wf *Workflow) Driver() WorkflowProcess {
	return wf.driver
}

// SetDriver sets the driver process of the workflow to the provided process
func (wf *Workflow) SetDriver(driver WorkflowProcess) {
	wf.driver = driver
}

// IncConcurrentTasks increases the conter for how many concurrent tasks are
// currently running in the workflow
func (wf *Workflow) IncConcurrentTasks(slots int) {
	// We must lock so that multiple processes don't end up with partially "filled slots"
	wf.concurrentTasksMx.Lock()
	for i := 0; i < slots; i++ {
		wf.concurrentTasks <- struct{}{}
		Debug.Println("Increased concurrent tasks")
	}
	wf.concurrentTasksMx.Unlock()
}

// DecConcurrentTasks decreases the conter for how many concurrent tasks are
// currently running in the workflow
func (wf *Workflow) DecConcurrentTasks(slots int) {
	for i := 0; i < slots; i++ {
		<-wf.concurrentTasks
		Debug.Println("Decreased concurrent tasks")
	}
}

// ConnectLast connects the last (most downstream) out-ports in the workflow to
// an implicit sink process which will be used to drive the workflow. This can
// be used instead of manually creating a sink, connecting it, and setting it
// as the driver process of the workflow.
func (wf *Workflow) ConnectLast(outPort *OutPort) {
	wf.sink.Connect(outPort)
	// Make sure the sink is also the driver
	wf.driver = wf.sink
}

func (wf *Workflow) readyToRun(procs ...WorkflowProcess) bool {
	if len(procs) == 0 {
		Error.Println(wf.name + ": The workflow is empty. Did you forget to add the processes to it?")
		return false
	}
	if wf.sink == nil {
		Error.Println(wf.name + ": sink is nil!")
		return false
	}
	for _, proc := range procs {
		if !proc.IsConnected() {
			Error.Println(wf.name + ": Not everything connected. Workflow shutting down.")
			return false
		}
	}
	return true
}

// RunProcs runs a specified set of processes only, with driverProc as the
// driver process
func (wf *Workflow) RunProcs(driverProc WorkflowProcess, procs ...WorkflowProcess) {
	if !wf.readyToRun(procs...) {
		Error.Fatalln("Workflow not ready to run, due to previously reported errors, so exiting.")
	}
	for _, proc := range procs {
		if proc != driverProc { // Don't start the driver process in background
			Debug.Printf(wf.name+": Starting process %s in new go-routine", proc.Name())
			go proc.Run()
		}
	}
	Debug.Printf(wf.name + ": Starting sink in main go-routine")
	driverProc.Run()
}

// RunProcsByName runs a specified set of processes only, specified by their
// names as strings, with driverProcName as the name for the driver process
func (wf *Workflow) RunProcsByName(driverProcName string, procNames ...string) {
	procs := []WorkflowProcess{}
	for _, procName := range procNames {
		procs = append(procs, wf.Proc(procName))
	}
	driverProc := wf.Proc(driverProcName)
	wf.RunProcs(driverProc, procs...)
}

// Run runs the workflow
func (wf *Workflow) Run() {
	procs := []WorkflowProcess{}
	for _, p := range wf.procs {
		procs = append(procs, p)
	}
	wf.RunProcs(wf.driver, procs...)
}
