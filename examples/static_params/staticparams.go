package main

import (
	// "fmt"
	sci "github.com/scipipe/scipipe"
)

func main() {
	sci.InitLogAudit()

	// Init
	fls := NewFileSender()

	params := map[string]string{"a": "a1", "b": "b1", "c": "c1"}

	abc := sci.ShellExpand("abc", "echo {p:a} {p:b} {p:c} > {o:out} # {i:in}", nil, nil, params)
	abc.PathFormatters["out"] = func(t *sci.SciTask) string {
		return t.GetInPath("in")
	}

	prt := sci.NewFromShell("prt", "echo {i:in} >> log.txt")

	// Connect
	abc.In["in"].Connect(fls.Out)
	prt.In["in"].Connect(abc.Out["out"])

	// Pipe it up
	pl := sci.NewPipelineRunner()
	pl.AddProcesses(fls, abc, prt)

	// Run
	pl.Run()
}

// --------------------------------
//         A Custom task
// --------------------------------

type FileSender struct {
	sci.Process
	Out *sci.FilePort
}

func NewFileSender() *FileSender {
	return &FileSender{
		Out: sci.NewFilePort(),
	}
}

func (proc *FileSender) Run() {
	defer proc.Out.Close()
	for _, fn := range []string{"file1.txt", "file2.txt", "file3.txt"} {
		proc.Out.Chan <- sci.NewFileTarget(fn)
	}
}

func (proc *FileSender) IsConnected() bool {
	return proc.Out.IsConnected()
}
