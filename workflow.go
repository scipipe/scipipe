package scipipe

import (
	"os"
)

// ----------------------------------------------------------------------------
// Workflow
// ----------------------------------------------------------------------------

type Workflow struct {
	name       string
	procs      map[string]Process
	driver     Process
	driverName string
}

func NewWorkflow(name string) *Workflow {
	if !LogExists {
		InitLogInfo()
	}
	return &Workflow{
		name:   name,
		procs:  map[string]Process{},
		driver: nil,
	}
}

func (wf *Workflow) Add(proc Process) {
	wf.procs[proc.Name()] = proc
}

func (wf *Workflow) NewFromShell(procName string, commandPattern string) *SciProcess {
	proc := NewFromShell(procName, commandPattern)
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

func (wf *Workflow) Driver() Process {
	return wf.driver
}

func (wf *Workflow) SetDriver(driver Process) {
	wf.driver = driver
	wf.driverName = driver.Name()
}

func (wf *Workflow) Run() {
	if len(wf.procs) == 0 {
		Error.Println(wf.name + ": The workflow is empty. Did you forget to add the processes to it?")
		os.Exit(1)
	}
	if wf.driver == nil {
		Error.Println(wf.name + ": No driver (process) added. Please set one, with wf.SetDriver()")
		os.Exit(1)
	}
	everythingConnected := true
	for _, proc := range wf.procs {
		if !proc.IsConnected() {
			everythingConnected = false
		}
	}
	if !everythingConnected {
		Error.Println(wf.name + ": Not everything connected. Workflow shutting down.")
		os.Exit(1)
	} else {
		for pname, proc := range wf.procs {
			Debug.Printf(wf.name+": Starting process %s in new go-routine", pname)
			go proc.Run()
		}
		Debug.Printf(wf.name + ": Starting driver process in main go-routine")
		wf.driver.Run()
	}
}
