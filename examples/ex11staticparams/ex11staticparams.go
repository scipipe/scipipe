package main

import (
	// "fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogAudit()
	// Init
	fls := NewFileSender()

	params := map[string]string{"a": "a1", "b": "b1", "c": "c1"}

	abc := sci.ShExp("echo {p:a} {p:b} {p:c} > {o:out} # {i:in}", nil, nil, params)
	abc.OutPathFuncs["out"] = func() string {
		return abc.GetInPath("in")
	}

	prt := sci.Sh("echo {i:in} >> log.txt")

	// Connect
	abc.InPorts["in"] = fls.Out
	prt.InPorts["in"] = abc.OutPorts["out"]

	// Pipe it up
	pl := sci.NewPipeline()
	pl.AddProcs(fls, abc, prt)

	// Run
	pl.Run()
}

// --------------------------------
//         A Custom task
// --------------------------------

type FileSender struct {
	sci.BaseProcess
	Out chan *sci.FileTarget
}

func NewFileSender() *FileSender {
	return &FileSender{
		Out: make(chan *sci.FileTarget, sci.BUFSIZE),
	}
}

func (proc *FileSender) Run() {
	defer close(proc.Out)
	for _, fn := range []string{"file1.txt", "file2.txt", "file3.txt"} {
		proc.Out <- sci.NewFileTarget(fn)
	}
}
