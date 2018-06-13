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
	abc.InParamPort("a").From(cmb.A())
	abc.InParamPort("b").From(cmb.B())
	abc.InParamPort("c").From(cmb.C())
	prt.In("in").From(abc.Out("out"))

	wf.Run()
}

type CombinatoricsGen struct {
	sci.BaseProcess
}

func NewCombinatoricsGen(wf *sci.Workflow, name string) *CombinatoricsGen {
	p := &CombinatoricsGen{
		BaseProcess: sci.NewBaseProcess(wf, name),
	}
	p.InitOutParamPort(p, "a")
	p.InitOutParamPort(p, "b")
	p.InitOutParamPort(p, "c")
	wf.AddProc(p)
	return p
}

func (p *CombinatoricsGen) A() *sci.OutParamPort { return p.OutParamPort("a") }
func (p *CombinatoricsGen) B() *sci.OutParamPort { return p.OutParamPort("b") }
func (p *CombinatoricsGen) C() *sci.OutParamPort { return p.OutParamPort("c") }

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
