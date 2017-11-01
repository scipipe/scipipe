package components

import (
	"github.com/scipipe/scipipe"
)

type MapToKeys struct {
	In       *scipipe.FilePort
	Out      *scipipe.FilePort
	procName string
	mapFunc  func(ip *scipipe.InformationPacket) map[string]string
}

func NewMapToKeys(wf *scipipe.Workflow, name string, mapFunc func(ip *scipipe.InformationPacket) map[string]string) *MapToKeys {
	mtp := &MapToKeys{
		procName: name,
		mapFunc:  mapFunc,
		In:       scipipe.NewFilePort(),
		Out:      scipipe.NewFilePort(),
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
