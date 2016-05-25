package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
	"runtime"
)

func main() {
	sci.InitLogDebug()

	runtime.GOMAXPROCS(4)
	cmb := NewCombinatoricsTask()

	// An abc file printer
	abc := sci.Shell("abc", "echo {p:a} {p:b} {p:c} > {o:out}; sleep 1")
	abc.Spawn = true
	abc.PathFormatters["out"] = func(t *sci.SciTask) string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			t.Params["a"],
			t.Params["b"],
			t.Params["c"],
		)
	}

	// A printer task
	prt := sci.Shell("printer", "cat {i:in} >> log.txt")
	prt.Spawn = false

	// Connection info
	abc.ParamPorts["a"] = cmb.A
	abc.ParamPorts["b"] = cmb.B
	abc.ParamPorts["c"] = cmb.C
	prt.InPorts["in"] = abc.OutPorts["out"]

	pl := sci.NewPipelineRunner()
	pl.AddProcesses(cmb, abc, prt)
	pl.Run()
}

type CombinatoricsTask struct {
	sci.Process
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

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				proc.A <- a
				proc.B <- b
				proc.C <- c
			}
		}
	}
}
