package components

import (
	"github.com/scipipe/scipipe"
)

type MapToKeys struct {
	In       *scipipe.InPort
	Out      *scipipe.OutPort
	procName string
	mapFunc  func(ip *scipipe.IP) map[string]string
}

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

func (p *MapToKeys) Name() string {
	return p.procName
}

func (p *MapToKeys) IsConnected() bool {
	return p.In.IsConnected() && p.Out.IsConnected()
}

func (p *MapToKeys) Run() {
	defer p.Out.Close()
	for ip := range p.In.Chan {
		newKeys := p.mapFunc(ip)
		ip.AddKeys(newKeys)
		ip.WriteAuditLogToFile()
		p.Out.Send(ip)
	}
}
