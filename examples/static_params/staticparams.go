package main

import (
	// "fmt"
	sci "github.com/scipipe/scipipe"
)

func main() {
	// Init
	wf := sci.NewWorkflow("staticparams_wf", 4)

	fls := NewFileSender("filesender")
	wf.AddProc(fls)

	params := map[string]string{"a": "a1", "b": "b1", "c": "c1"}

	abc := sci.ShellExpand(wf, "abc", "echo {p:a} {p:b} {p:c} > {o:out} # {i:in}", nil, nil, params)
	abc.SetPathCustom("out", func(task *sci.SciTask) string {
		return task.InPath("in")
	})

	prt := wf.NewProc("prt", "echo {i:in} >> log.txt")

	// Connect
	abc.In("in").Connect(fls.Out)
	prt.In("in").Connect(abc.Out("out"))

	// Pipe it up
	wf.SetDriver(prt)

	// Run
	wf.Run()
}

// --------------------------------
//         A Custom task
// --------------------------------

type FileSender struct {
	name string
	Out  *sci.OutPort
}

func NewFileSender(name string) *FileSender {
	return &FileSender{
		name: name,
		Out:  sci.NewOutPort(),
	}
}

func (p *FileSender) Run() {
	defer p.Out.Close()
	for _, fn := range []string{"file1.txt", "file2.txt", "file3.txt"} {
		p.Out.Send(sci.NewIP(fn))
	}
}

func (p *FileSender) Name() string {
	return p.name
}

func (p *FileSender) IsConnected() bool {
	return p.Out.IsConnected()
}
