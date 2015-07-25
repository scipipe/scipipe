package scipipe

import (
	"fmt"
	"reflect"
)

type Pipeline struct {
	processes []process
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Short-hand method
func (pl *Pipeline) AddProc(proc process) {
	pl.AddProcess(proc)
}
func (pl *Pipeline) AddProcess(proc process) {
	pl.processes = append(pl.processes, proc)
}

// Short-hand method
func (pl *Pipeline) AddProcs(procs ...process) {
	for _, proc := range procs {
		pl.AddProcess(proc)
	}
}
func (pl *Pipeline) AddProcesses(procs ...process) {
	for _, proc := range procs {
		pl.AddProcess(proc)
	}
}

func (pl *Pipeline) PrintProcesses() {
	for i, proc := range pl.processes {
		fmt.Printf("Process %d: %v\n", i, reflect.TypeOf(proc))
	}
}

func (pl *Pipeline) Run() {
	for i, proc := range pl.processes {
		if i < len(pl.processes)-1 {
			go proc.Run()
		} else {
			proc.Run()
		}
	}
}
