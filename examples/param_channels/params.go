package main

import (
	"fmt"
	"runtime"

	sci "github.com/scipipe/scipipe"
)

func main() {
	runtime.GOMAXPROCS(4)
	wf := sci.NewWorkflow("test_wf", 4)

	cmb := NewCombinatoricsGen(wf, "combgen")

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
	abc.ParamInPort("a").Connect(cmb.A())
	abc.ParamInPort("b").Connect(cmb.B())
	abc.ParamInPort("c").Connect(cmb.C())
	prt.In("in").Connect(abc.Out("out"))

	wf.Run()
}

type CombinatoricsGen struct {
	sci.BaseProcess
}

func NewCombinatoricsGen(wf *sci.Workflow, name string) *CombinatoricsGen {
	p := &CombinatoricsGen{
		BaseProcess: sci.NewBaseProcess(wf, name),
	}
	p.InitParamInPort(p, "a")
	p.InitParamInPort(p, "b")
	p.InitParamInPort(p, "c")
	wf.AddProc(p)
	return p
}

func (p *CombinatoricsGen) A() *sci.ParamOutPort { return p.ParamOutPort("a") }
func (p *CombinatoricsGen) B() *sci.ParamOutPort { return p.ParamOutPort("b") }
func (p *CombinatoricsGen) C() *sci.ParamOutPort { return p.ParamOutPort("c") }

func (p *CombinatoricsGen) Run() {
	defer p.CloseAllOutPorts()

	for _, a := range []string{"a1", "a2", "a3"} {
		for _, b := range []string{"b1", "b2", "b3"} {
			for _, c := range []string{"c1", "c2", "c3"} {
				p.A().Send(a)
				p.B().Send(b)
				p.C().Send(c)
			}
		}
	}
}
