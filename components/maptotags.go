package components

import (
	"github.com/scipipe/scipipe"
)

// MapToTags is a process that runs a function provided by the user, upon
// initialization, that will provide a map of tag:value pairs, based in IPs read
// on the In-port. The tag:value pairs (maps) are added to the IPs on the
// out-port, which are identical to the incoming IPs, except for the new
// tag:value map
type MapToTags struct {
	scipipe.BaseProcess
	mapFunc func(ip *scipipe.FileIP) map[string]string
}

// NewMapToTags returns an initialized MapToTags process
func NewMapToTags(wf *scipipe.Workflow, name string, mapFunc func(ip *scipipe.FileIP) map[string]string) *MapToTags {
	p := &MapToTags{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		mapFunc:     mapFunc,
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// In takes input files the content of which the map function will be run,
// to generate tags
func (p *MapToTags) In() *scipipe.InPort { return p.InPort("in") }

// Out outputs files that are supplemented with tags by the map function.
func (p *MapToTags) Out() *scipipe.OutPort { return p.OutPort("out") }

// Run runs the MapToTags process
func (p *MapToTags) Run() {
	defer p.CloseAllOutPorts()
	for ip := range p.In().Chan {
		newTags := p.mapFunc(ip)
		ip.AddTags(newTags)
		ip.WriteAuditLogToFile("")
		p.Out().Send(ip)
	}
}
