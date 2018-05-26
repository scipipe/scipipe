// Package scipipe is a library for writing scientific workflows (sometimes
// also called "pipelines") of shell commands that depend on each other, in the
// Go programming languages. It was initially designed for problems in
// cheminformatics and bioinformatics, but should apply equally well to any
// domain involving complex pipelines of interdependent shell commands.
package scipipe

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
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
	logFile           string
}

// WorkflowProcess is an interface for processes to be handled by Workflow
type WorkflowProcess interface {
	Name() string
	InPorts() map[string]*InPort
	OutPorts() map[string]*OutPort
	ParamInPorts() map[string]*ParamInPort
	ParamOutPorts() map[string]*ParamOutPort
	Connected() bool
	Run()
}

// ----------------------------------------------------------------------------
// Factory function(s)
// ----------------------------------------------------------------------------

// NewWorkflow returns a new Workflow
func NewWorkflow(name string, maxConcurrentTasks int) *Workflow {
	wf := newWorkflowWithoutLogging(name, maxConcurrentTasks)

	// Set up logging
	allowedCharsPtrn, err := regexp.Compile("[^a-z0-9_]")
	if err != nil {
		fmt.Println("Could not compile regex for workflow name")
		os.Exit(1)
	}
	wfNameNormalized := allowedCharsPtrn.ReplaceAllString(strings.ToLower(name), "-")
	wf.logFile = "log/scipipe-" + time.Now().Format("20060102-150405") + "-" + wfNameNormalized + ".log"
	InitLogAuditToFile(wf.logFile)

	return wf
}

// NewWorkflowCustomLogFile returns a new Workflow, with
func NewWorkflowCustomLogFile(name string, maxConcurrentTasks int, logFile string) *Workflow {
	wf := newWorkflowWithoutLogging(name, maxConcurrentTasks)

	wf.logFile = logFile
	InitLogAuditToFile(logFile)

	return wf
}

func newWorkflowWithoutLogging(name string, maxConcurrentTasks int) *Workflow {
	wf := &Workflow{
		name:            name,
		procs:           map[string]WorkflowProcess{},
		concurrentTasks: make(chan struct{}, maxConcurrentTasks),
	}
	sink := NewSink(wf, name+"_default_sink")
	wf.sink = sink
	wf.driver = sink
	return wf
}

// ----------------------------------------------------------------------------
// Main API methods
// ----------------------------------------------------------------------------

// Name returns the name of the workflow
func (wf *Workflow) Name() string {
	return wf.name
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
	if _, ok := wf.procs[procName]; !ok {
		Failf("No process named '%s' in workflow '%s'", procName, wf.Name())
	}
	return wf.procs[procName]
}

// Procs returns a map of all processes keyed by their names in the workflow
func (wf *Workflow) Procs() map[string]WorkflowProcess {
	return wf.procs
}

// AddProc adds a Process to the workflow, to be run when the workflow runs
func (wf *Workflow) AddProc(proc WorkflowProcess) {
	if wf.procs[proc.Name()] != nil {
		Failf(wf.name+" workflow: A process with name '%s' already exists in the workflow! Use a more unique name!\n", proc.Name())
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

// Sink returns the sink process of the workflow
func (wf *Workflow) Sink() *Sink {
	return wf.sink
}

// SetSink sets the sink of the workflow to the provided sink process
func (wf *Workflow) SetSink(sink *Sink) {
	if wf.sink.Connected() {
		Fail("Trying to replace a sink which is already connected. Are you combining SetSink() with ConnectFinalOutPort()? That is not allowed!")
	}
	wf.sink = sink
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

// ----------------------------------------------------------------------------
// Run methods
// ----------------------------------------------------------------------------

// Run runs all the processes of the workflow
func (wf *Workflow) Run() {
	wf.runProcs(wf.procs)
}

// RunTo runs all processes upstream of, and including, the process with
// names provided as arguments
func (wf *Workflow) RunTo(finalProcNames ...string) {
	procs := []WorkflowProcess{}
	for _, procName := range finalProcNames {
		procs = append(procs, wf.Proc(procName))
	}
	wf.RunToProcs(procs...)
}

// RunToRegex runs all processes upstream of, and including, the process
// whose name matches any of the provided regexp patterns
func (wf *Workflow) RunToRegex(procNamePatterns ...string) {
	procsToRun := []WorkflowProcess{}
	for _, pattern := range procNamePatterns {
		regexpPtrn, err := regexp.Compile(pattern)
		CheckWithMsg(err, fmt.Sprintf("Regex pattern doesn't work: %s", pattern))
		for procName, proc := range wf.Procs() {
			matches := regexpPtrn.MatchString(procName)
			if matches {
				procsToRun = append(procsToRun, proc)
			}
		}
	}
	wf.RunToProcs(procsToRun...)
}

// RunToProcs runs all processes upstream of, and including, the process strucs
// provided as arguments
func (wf *Workflow) RunToProcs(finalProcs ...WorkflowProcess) {
	procsToRun := map[string]WorkflowProcess{}
	for _, finalProc := range finalProcs {
		procsToRun = mergeWFMaps(procsToRun, upstreamProcsForProc(finalProc))
		procsToRun[finalProc.Name()] = finalProc
	}
	wf.runProcs(procsToRun)
}

// ----------------------------------------------------------------------------
// Helper methods for running the workflow
// ----------------------------------------------------------------------------

// runProcs runs a specified set of processes only
func (wf *Workflow) runProcs(procs map[string]WorkflowProcess) {
	wf.reconnectDeadEndConnections(procs)

	if !wf.readyToRun(procs) {
		Fail("Workflow not ready to run, due to previously reported errors, so exiting.")
	}

	for _, proc := range procs {
		Debug.Printf(wf.name+": Starting process %s in new go-routine", proc.Name())
		go proc.Run()
	}

	Debug.Printf(wf.name + ": Starting driver process in main go-routine")
	Audit.Printf("| workflow:%-23s | Starting workflow (Writing log to %s)", wf.Name(), wf.logFile)
	wf.driver.Run()
	Audit.Printf("| workflow:%-23s | Finished worklfow (Log written to %s)", wf.Name(), wf.logFile)
}

func (wf *Workflow) readyToRun(procs map[string]WorkflowProcess) bool {
	if len(procs) == 0 {
		Error.Println(wf.name + ": The workflow is empty. Did you forget to add the processes to it?")
		return false
	}
	if wf.sink == nil {
		Error.Println(wf.name + ": sink is nil!")
		return false
	}
	for _, proc := range procs {
		if !proc.Connected() {
			Error.Println(wf.name + ": Not everything connected. Workflow shutting down.")
			return false
		}
	}
	return true
}

// reconnectDeadEndConnections disonnects connections to processes which are not
// in the set of processes to be run, and, if an out-port for a process that is supposed to be
// run becomes disconnected, it is connected to the sink instead
func (wf *Workflow) reconnectDeadEndConnections(procs map[string]WorkflowProcess) {
	foundNewDriverProc := false

	for pname, proc := range procs {
		// OutPorts
		for _, opt := range proc.OutPorts() {
			for iptName, ipt := range opt.RemotePorts {
				// If the remotely connected process is not among the ones to run ...
				if ipt.Process() == nil {
					Debug.Printf("Disconnecting in-port %s from out-port %s", ipt.Name(), opt.Name())
					opt.Disconnect(iptName)
				} else if _, ok := procs[ipt.Process().Name()]; !ok {
					Debug.Printf("Disconnecting in-port %s from out-port %s", ipt.Name(), opt.Name())
					opt.Disconnect(iptName)
				}
			}
			if !opt.Connected() {
				Debug.Printf("Connecting disconnected out-port %s of process %s to workflow sink", opt.Name(), opt.Process().Name())
				wf.sink.Connect(opt)
			}
		}

		// ParamOutPorts
		for _, pop := range proc.ParamOutPorts() {
			for rppName, rpp := range pop.RemotePorts {
				// If the remotely connected process is not among the ones to run ...
				if rpp.Process() == nil {
					Debug.Printf("Disconnecting in-port %s from out-port %s", rpp.Name(), pop.Name())
					pop.Disconnect(rppName)
				} else if _, ok := procs[rpp.Process().Name()]; !ok {
					Debug.Printf("Disconnecting in-port %s from out-port %s", rpp.Name(), pop.Name())
					pop.Disconnect(rppName)
				}
			}
			if !pop.Connected() {
				Debug.Printf("Connecting disconnected out-port %s of process %s to workflow sink", pop.Name(), pop.Process().Name())
				wf.sink.ConnectParam(pop)
			}
		}

		if len(proc.OutPorts()) == 0 && len(proc.ParamOutPorts()) == 0 {
			if foundNewDriverProc {
				Failf("Found more than one process without out-ports nor out-param ports. Cannot use both as drivers (One of them being '%s'). Adapt your workflow accordingly.", proc.Name())
			}
			foundNewDriverProc = true
			wf.driver = proc
			delete(wf.procs, pname) // A process can't both be the driver, and be included in the main procs map
		}
	}
}

// upstreamProcsForProc returns all processes it is connected to, either
// directly or indirectly, via its in-ports and param-in-ports
func upstreamProcsForProc(proc WorkflowProcess) map[string]WorkflowProcess {
	procs := map[string]WorkflowProcess{}
	for _, inp := range proc.InPorts() {
		for _, rpt := range inp.RemotePorts {
			procs[rpt.Process().Name()] = rpt.Process()
			mergeWFMaps(procs, upstreamProcsForProc(rpt.Process()))
		}
	}
	for _, pip := range proc.ParamInPorts() {
		for _, rpp := range pip.RemotePorts {
			procs[rpp.Process().Name()] = rpp.Process()
			mergeWFMaps(procs, upstreamProcsForProc(rpp.Process()))
		}
	}
	return procs
}

func mergeWFMaps(a map[string]WorkflowProcess, b map[string]WorkflowProcess) map[string]WorkflowProcess {
	for k, v := range b {
		a[k] = v
	}
	return a
}
