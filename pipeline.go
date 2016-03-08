package scipipe

import (
	"fmt"
	"reflect"
)

type PipelineRunner struct {
	processes []Process
}

func NewPipelineRunner() *PipelineRunner {
	return &PipelineRunner{}
}

// Short-hand method
func (pl *PipelineRunner) AddProc(proc Process) {
	pl.AddProcess(proc)
}
func (pl *PipelineRunner) AddProcess(proc Process) {
	pl.processes = append(pl.processes, proc)
}

// Short-hand method
func (pl *PipelineRunner) AddProcs(procs ...Process) {
	for _, proc := range procs {
		pl.AddProcess(proc)
	}
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
