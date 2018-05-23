package components

import (
	"github.com/scipipe/scipipe"
)

// ParamSource will feed parameters on an out-port
type ParamSource struct {
	scipipe.BaseProcess
	params []string
}

// NewParamSource returns a new ParamSource
func NewParamSource(wf *scipipe.Workflow, name string, params ...string) *ParamSource {
	p := &ParamSource{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		params:      params,
	}
	p.InitParamOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port, on which parameters the process was initialized
// with, will be retrieved.
func (p *ParamSource) Out() *scipipe.ParamOutPort { return p.ParamOutPort("out") }

// Run runs the process
func (p *ParamSource) Run() {
	defer p.CloseAllOutPorts()
	for _, param := range p.params {
		p.Out().Send(param)
	}
}
