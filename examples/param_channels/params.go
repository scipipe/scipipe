package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(4)
	cmb := NewCombinatoricsGen("combgen")

	// An abc file printer
	abc := sci.NewProc("abc", "echo {p:a} {p:b} {p:c} > {o:out}; sleep 1")
	abc.Spawn = true
	abc.SetPathCustom("out", func(t *sci.SciTask) string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			t.Params["a"],
			t.Params["b"],
			t.Params["c"],
		)
	})

	// A printer task
	prt := sci.NewProc("printer", "cat {i:in} >> log.txt")
	prt.Spawn = false

	// Connection info
	abc.ParamPort("a").Connect(cmb.A)
	abc.ParamPort("b").Connect(cmb.B)
	abc.ParamPort("c").Connect(cmb.C)
	prt.In("in").Connect(abc.Out("out"))

	wf := sci.NewWorkflow("combwf")
	wf.AddProcs(cmb, abc)
	wf.SetDriver(prt)
	wf.Run()
}

type CombinatoricsGen struct {
	sci.Process
	name string
	A    *sci.ParamPort
	B    *sci.ParamPort
	C    *sci.ParamPort
}

func NewCombinatoricsGen(name string) *CombinatoricsGen {
	return &CombinatoricsGen{
		name: name,
		A:    sci.NewParamPort(),
		B:    sci.NewParamPort(),
		C:    sci.NewParamPort(),
	}
}

func (p *CombinatoricsGen) Name() string {
	return p.name
}

func (p *CombinatoricsGen) Run() {
	defer p.A.Close()
	defer p.B.Close()
	defer p.C.Close()

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				p.A.Send(a)
				p.B.Send(b)
				p.C.Send(c)
			}
		}
	}
}

func (p *CombinatoricsGen) IsConnected() bool { return true }
