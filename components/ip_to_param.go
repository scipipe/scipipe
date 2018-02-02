package components

import (
	"strings"

	"github.com/scipipe/scipipe"
)

// IPToParamConverter takes a file target on its FilePath in-port, reads its
// content (assuming a single value), removing any newlines, spaces or tabs,
// and sends the value on the OutParam parameter port.
type IPToParamConverter struct {
	name     string
	InFile   *scipipe.InPort
	OutParam *scipipe.ParamPort
}

// NewIPToParamConverter instantiates a new IPToParamConverter
func NewIPToParamConverter(wf *scipipe.Workflow, name string) *IPToParamConverter {
	p := &IPToParamConverter{
		name:     name,
		InFile:   scipipe.NewInPort("in_file"),
		OutParam: scipipe.NewParamPort(),
	}
	wf.AddProc(p)
	return p
}

// Name returns the name of the IPToParamConverter process
func (p *IPToParamConverter) Name() string {
	return p.name
}

// IsConnected tells whether all the ports of the IPToParamConverter process are
// connected
func (p *IPToParamConverter) IsConnected() bool {
	return p.InFile.IsConnected() && p.OutParam.IsConnected()
}

// Run the IPToParamConverter
func (p *IPToParamConverter) Run() {
	defer p.OutParam.Close()

	for ip := range p.InFile.Chan {
		s := string(ip.Read())
		s = strings.Trim(s, " \r\n\t")
		p.OutParam.Send(s)
	}
}
