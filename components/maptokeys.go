package components

import (
	"github.com/scipipe/scipipe"
)

type MapToKeys struct {
	In       *scipipe.Port
	Out      *scipipe.Port
	procName string
	mapFunc  func(ip *scipipe.IP) map[string]string
}

func NewMapToKeys(wf *scipipe.Workflow, name string, mapFunc func(ip *scipipe.IP) map[string]string) *MapToKeys {
	mtp := &MapToKeys{
		procName: name,
		mapFunc:  mapFunc,
		In:       scipipe.NewPort(),
		Out:      scipipe.NewPort(),
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
	go p.In.RunMergeInputs()
	for ip := range p.In.InChan {
		newKeys := p.mapFunc(ip)
		ip.AddKeys(newKeys)
		ip.WriteAuditLogToFile()
		p.Out.Send(ip)
	}
}
