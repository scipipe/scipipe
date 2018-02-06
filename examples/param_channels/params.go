package main

import (
	"fmt"
	"runtime"

	sci "github.com/scipipe/scipipe"
)

func main() {
	runtime.GOMAXPROCS(4)
	wf := sci.NewWorkflow("test_wf", 4)

	cmb := NewCombinatoricsGen("combgen")
	wf.AddProc(cmb)

	// An abc file printer
	abc := wf.NewProc("abc", "echo {p:a} {p:b} {p:c} > {o:out}; sleep 1")
	abc.Spawn = true
	abc.SetPathCustom("out", func(t *sci.Task) string {
		return fmt.Sprintf(
			"%s_%s_%s.txt",
			t.Param("a"),
			t.Param("b"),
			t.Param("c"),
		)
	})

	// A printer task
	prt := wf.NewProc("printer", "cat {i:in} >> log.txt")
	prt.Spawn = false

	// Connection info
	abc.ParamInPort("a").Connect(cmb.A)
	abc.ParamInPort("b").Connect(cmb.B)
	abc.ParamInPort("c").Connect(cmb.C)
	prt.In("in").Connect(abc.Out("out"))

	wf.Run()
}

type CombinatoricsGen struct {
	sci.EmptyWorkflowProcess
	name string
	A    *sci.ParamOutPort
	B    *sci.ParamOutPort
	C    *sci.ParamOutPort
}

func NewCombinatoricsGen(name string) *CombinatoricsGen {
	cmb := &CombinatoricsGen{
		name: name,
		A:    sci.NewParamOutPort("a"),
		B:    sci.NewParamOutPort("b"),
		C:    sci.NewParamOutPort("c"),
	}
	cmb.A.SetProcess(cmb)
	cmb.B.SetProcess(cmb)
	cmb.C.SetProcess(cmb)
	return cmb
}

// ParamOutPorts is a required interface method (WorkflowProcess)
func (p *CombinatoricsGen) ParamOutPorts() map[string]*sci.ParamOutPort {
	return map[string]*sci.ParamOutPort{
		p.A.Name(): p.A,
		p.B.Name(): p.B,
		p.C.Name(): p.C,
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

func (p *CombinatoricsGen) Connected() bool { return true }
