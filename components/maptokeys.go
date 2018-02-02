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
	In       *scipipe.InPort
	Out      *scipipe.OutPort
	procName string
	mapFunc  func(ip *scipipe.IP) map[string]string
}

// NewMapToKeys returns an initialized MapToKeys process
func NewMapToKeys(wf *scipipe.Workflow, name string, mapFunc func(ip *scipipe.IP) map[string]string) *MapToKeys {
	mtp := &MapToKeys{
		procName: name,
		mapFunc:  mapFunc,
		In:       scipipe.NewInPort("in"),
		Out:      scipipe.NewOutPort("out"),
	}
	wf.AddProc(mtp)
	return mtp
}

// Name returns the name of the MapToKeys process
func (p *MapToKeys) Name() string {
	return p.procName
}

// IsConnected tells whether all ports of the MapToKeys process are connected
func (p *MapToKeys) IsConnected() bool {
	return p.In.IsConnected() && p.Out.IsConnected()
}

// Run runs the MapToKeys process
func (p *MapToKeys) Run() {
	defer p.Out.Close()
	for ip := range p.In.Chan {
		newKeys := p.mapFunc(ip)
		ip.AddKeys(newKeys)
		ip.WriteAuditLogToFile()
		p.Out.Send(ip)
	}
}
