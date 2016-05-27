package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
	"runtime"
)

func main() {
	sci.InitLogAudit()

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
	abc.ParamPorts["a"].Connect(cmb.A)
	abc.ParamPorts["b"].Connect(cmb.B)
	abc.ParamPorts["c"].Connect(cmb.C)
	prt.InPorts["in"].Connect(abc.OutPorts["out"])

	pl := sci.NewPipelineRunner()
	pl.AddProcesses(cmb, abc, prt)
	pl.Run()
}

type CombinatoricsTask struct {
	sci.Process
	A *sci.ParamPort
	B *sci.ParamPort
	C *sci.ParamPort
}

func NewCombinatoricsTask() *CombinatoricsTask {
	return &CombinatoricsTask{
		A: sci.NewParamPort(),
		B: sci.NewParamPort(),
		C: sci.NewParamPort(),
	}
}

func (proc *CombinatoricsTask) Run() {
	defer proc.A.Close()
	defer proc.B.Close()
	defer proc.C.Close()

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				proc.A.Chan <- a
				proc.B.Chan <- b
				proc.C.Chan <- c
			}
		}
	}
}

func (proc *CombinatoricsTask) IsConnected() bool { return true }
