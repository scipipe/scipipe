package components

import (
	"github.com/scipipe/scipipe"
	"sync"
)

type Merger struct {
	ins [](*scipipe.FilePort)
	Out *scipipe.FilePort
}

func NewMerger() *Merger {
	return &Merger{}
}

func (proc *Merger) Run() {
	var wg sync.WaitGroup
	wg.Add(len(proc.ins))
	for _, inp := range proc.ins {
		go func(ch chan *scipipe.InformationPacket) {
			for ft := range ch {
				proc.Out.Chan <- ft
			}
			wg.Done()
		}(inp.Chan)
	}
	wg.Wait()
	proc.Out.Close()
}

func (proc *Merger) Connect(outPort *scipipe.FilePort) {
	inPort := scipipe.NewFilePort()
	inPort.Connect(outPort)
	proc.ins = append(proc.ins, inPort)
}
