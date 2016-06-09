package scipipe

import (
	"fmt"
	"os"
	"reflect"
)

type PipelineRunner struct {
	processes []Process
}

func NewPipelineRunner() *PipelineRunner {
	return &PipelineRunner{}
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
		InitLogAudit()
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
