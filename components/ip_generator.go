package components

import "github.com/scipipe/scipipe"

// IPGenerator is initialized by a set of strings with file paths, and from that will
// return instantiated (generated) IP on its Out-port, when run.
type IPGenerator struct {
	scipipe.BaseProcess
	FilePaths []string
}

// NewIPGenerator initializes a new IPGenerator component from a list of file paths
func NewIPGenerator(wf *scipipe.Workflow, name string, filePaths ...string) (p *IPGenerator) {
	p = &IPGenerator{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		FilePaths:   filePaths,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port of the IPGenerator
func (p *IPGenerator) Out() *scipipe.OutPort {
	return p.OutPort("out")
}

// Run runs the IPGenerator process, returning instantiated IP
func (p *IPGenerator) Run() {
	defer p.Out().Close()
	for _, fp := range p.FilePaths {
		p.Out().Send(scipipe.NewIP(fp))
	}
}
