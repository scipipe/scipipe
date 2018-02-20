package components

import (
	"github.com/scipipe/scipipe"
)

// MapToKeys is a process that runs a function provided by the user, upon
// initialization, that will provide a map of key:value pairs, based in IPs read
// on the In-port. The key:value pairs (maps) are added to the IPs on the
// out-port, which are identical to the incoming IPs, except for the new
// key:value map
type MapToKeys struct {
	scipipe.BaseProcess
	mapFunc func(ip *scipipe.FileIP) map[string]string
}

// NewMapToKeys returns an initialized MapToKeys process
func NewMapToKeys(wf *scipipe.Workflow, name string, mapFunc func(ip *scipipe.FileIP) map[string]string) *MapToKeys {
	p := &MapToKeys{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		mapFunc:     mapFunc,
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// In takes input files the content of which the map function will be run,
// to generate keys
func (p *MapToKeys) In() *scipipe.InPort { return p.InPort("in") }

// Out outputs files that are supplemented with keys by the map function.
func (p *MapToKeys) Out() *scipipe.OutPort { return p.OutPort("out") }

// Run runs the MapToKeys process
func (p *MapToKeys) Run() {
	defer p.CloseAllOutPorts()
	for ip := range p.In().Chan {
		newKeys := p.mapFunc(ip)
		ip.AddKeys(newKeys)
		ip.WriteAuditLogToFile()
		p.Out().Send(ip)
	}
}
