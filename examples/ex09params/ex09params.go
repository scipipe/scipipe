package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
	"runtime"
)

func main() {
	sci.InitLogDebug()

	runtime.GOMAXPROCS(4)
	cmb := NewCombinatoricsTask()

	// An abc file printer
	abc := sci.Sh("echo {p:a} {p:b} {p:c} > {o:out}; sleep 1")
	abc.Spawn = true
	abc.OutPathFuncs["out"] = func() string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			abc.Params["a"],
			abc.Params["b"],
			abc.Params["c"],
		)
	}

	// A printer task
	prt := sci.Sh("cat {i:in} >> log.txt")
	prt.Spawn = false

	// Connection info
	abc.ParamPorts["a"] = cmb.A
	abc.ParamPorts["b"] = cmb.B
	abc.ParamPorts["c"] = cmb.C
	prt.InPorts["in"] = abc.OutPorts["out"]

	pl := sci.NewPipeline()
	pl.AddTasks(cmb, abc, prt)
	pl.Run()
}

type CombinatoricsTask struct {
	sci.BaseTask
	A chan string
	B chan string
	C chan string
}

func NewCombinatoricsTask() *CombinatoricsTask {
	return &CombinatoricsTask{
		A: make(chan string, sci.BUFSIZE),
		B: make(chan string, sci.BUFSIZE),
		C: make(chan string, sci.BUFSIZE),
	}
}

func (proc *CombinatoricsTask) Run() {
	defer close(proc.A)
	defer close(proc.B)
	defer close(proc.C)

	for _, a := range sci.SS("a1", "a2", "a3") {
		for _, b := range sci.SS("b1", "b2", "b3") {
			for _, c := range sci.SS("c1", "c2", "c3") {
				proc.A <- a
				proc.B <- b
				proc.C <- c
			}
		}
	}
}
