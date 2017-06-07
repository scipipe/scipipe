package scipipe

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// ----------------------------------------------------------------------------
// Pipeline
// ----------------------------------------------------------------------------

type Pipeline struct {
	Runner      *PipelineRunner
	processes   map[string]ShellProcess
	lastProcess ShellProcess
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		Runner:      &PipelineRunner{},
		processes:   map[string]ShellProcess{},
		lastProcess: nil,
	}
}

func (pl *Pipeline) AddProc(procName string, proc ShellProcess) {
	pl.processes[procName] = proc
	pl.Runner.AddProcess(proc)
	pl.lastProcess = proc
}

func (pl *Pipeline) GetProc(procName string) ShellProcess {
	return pl.processes[procName]
}

func (pl *Pipeline) GetProcs() map[string]ShellProcess {
	return pl.processes
}

func (pl *Pipeline) GetLastProc() ShellProcess {
	return pl.lastProcess
}

func (pl *Pipeline) NewProc(procName string, commandPattern string) {
	pl.AddProc(procName, NewFromShell(procName, commandPattern))
}

func (pl *Pipeline) SetPath(portSpec string, pathSpec string) {

}

func (pl *Pipeline) Connect(connSpec string) {
	directionLeft := true
	var bits []string
	if strings.Contains(connSpec, "<-") {
		bits = strings.Split(connSpec, "<-")
	} else if strings.Contains(connSpec, "->") {
		bits = strings.Split(connSpec, "->")
		directionLeft = false
	} else {
		Error.Println("Pipeline: No <- or -> in connection string: ", connSpec)
		os.Exit(1)
	}
	// Trim witespace
	for i, bit := range bits {
		bits[i] = strings.Trim(bit, " ")
	}
	part1 := bits[0]
	part2 := bits[1]

	var proc1Name string
	var proc2Name string
	var port1Name string
	var port2Name string

	if strings.Contains(part1, ".") {
		part1bits := strings.Split(part1, ".")
		proc1Name = part1bits[0]
		port1Name = part1bits[1]
	} else {
		Error.Println("Pipeline: No dot to separate process and port, in left part of connection string: ", connSpec)
		os.Exit(1)
	}

	if strings.Contains(part2, ".") {
		part2bits := strings.Split(part2, ".")
		proc2Name = part2bits[0]
		port2Name = part2bits[1]
	} else {
		Error.Println("Pipeline: No dot to separate process and port, in right part of connection string: ", connSpec)
		os.Exit(1)
	}

	if directionLeft {
		Connect(
			pl.GetProc(proc1Name).In(port1Name),
			pl.GetProc(proc2Name).Out(port2Name))
	} else {
		Connect(
			pl.GetProc(proc1Name).Out(port1Name), // <- Order of In/Out ports
			pl.GetProc(proc2Name).In(port2Name))  // <- switched here
	}
}

func (pl *Pipeline) Run() {
	sink := NewSink()
	pl.Runner.AddProcess(sink)

	lastProc := pl.GetLastProc()
	for _, port := range lastProc.GetOutPorts() {
		sink.Connect(port)
	}

	pl.Runner.Run()
}

// ----------------------------------------------------------------------------
// Pipeline Runner
// ----------------------------------------------------------------------------

type PipelineRunner struct {
	processes []Process
}

func NewPipelineRunner() *PipelineRunner {
	return &PipelineRunner{}
}

func (pl *PipelineRunner) NewFromShell(procName string, commandPattern string) *SciProcess {
	proc := NewFromShell(procName, commandPattern)
	pl.AddProcess(proc)
	return proc
}

func (pl *PipelineRunner) AddProcess(proc Process) {
	pl.processes = append(pl.processes, proc)
}

func (pl *PipelineRunner) AddProcesses(procs ...Process) {
	for _, proc := range procs {
		pl.AddProcess(proc)
	}
}

func (pl *PipelineRunner) PrintProcesses() {
	for i, proc := range pl.processes {
		fmt.Printf("Process %d: %v\n", i, reflect.TypeOf(proc))
	}
}

func (pl *PipelineRunner) Run() {
	if !LogExists {
		InitLogInfo()
	}
	if len(pl.processes) == 0 {
		Error.Println("PipelineRunner: The PipelineRunner is empty. Did you forget to add the processes to it?")
		os.Exit(1)
	}
	everythingConnected := true
	for _, proc := range pl.processes {
		if !proc.IsConnected() {
			everythingConnected = false
		}
	}
	if !everythingConnected {
		Error.Println("PipelineRunner: Pipeline shutting down, since not all ports are connected!")
		os.Exit(1)
	} else {
		for i, proc := range pl.processes {
			Debug.Printf("PipelineRunner: Looping over process %d: %v ...\n", i, proc)
			if i < len(pl.processes)-1 {
				Debug.Printf("PipelineRunner: Starting process %d in new go-routine: %v\n", i, proc)
				go proc.Run()
			} else {
				Debug.Printf("PipelineRunner: Starting process %d: in main go-routine: %v\n", i, proc)
				proc.Run()
			}
		}
	}
}
