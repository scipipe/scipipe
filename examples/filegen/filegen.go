package main

import (
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
)

func main() {
	wf := sp.NewWorkflow("filegenwf", 4)

	fq := spc.NewFileSource(wf, "file_src", "hej1.txt", "hej2.txt", "hej3.txt")

	fw := sp.NewProc(wf, "filewriter", "echo {i:in} > {o:out}")
	fw.SetOut("out", "{i:in}")
	fw.In("in").From(fq.Out())

	wf.Run()
}

// --------------------------------------------------------------------------------
// FileIPGenerator helper process
// --------------------------------------------------------------------------------

// FileIPGenerator is initialized by a set of strings with file paths, and from that will
// return instantiated (generated) FileIP on its Out-port, when run.
type FileIPGenerator struct {
	sp.BaseProcess
	FilePaths []string
}

// NewFileIPGenerator initializes a new FileIPGenerator component from a list of file paths
func NewFileIPGenerator(wf *sp.Workflow, name string, filePaths ...string) (p *FileIPGenerator) {
	p = &FileIPGenerator{
		BaseProcess: sp.NewBaseProcess(wf, name),
		FilePaths:   filePaths,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port of the FileIPGenerator
func (p *FileIPGenerator) Out() *sp.OutPort {
	return p.OutPort("out")
}

// Run runs the FileIPGenerator process, returning instantiated FileIP
func (p *FileIPGenerator) Run() {
	defer p.Out().Close()
	for _, fp := range p.FilePaths {
		p.Out().Send(sp.NewFileIP(fp))
	}
}
