package proclib

import (
	"github.com/scipipe/scipipe"
	"sync"
)

type Merger struct {
	ins [](*scipipe.InPort)
	Out *scipipe.OutPort
}

func NewMerger() *Merger {
	return &Merger{}
}

func (proc *Merger) Run() {
	var wg sync.WaitGroup
	wg.Add(len(proc.ins))
	for _, inp := range proc.ins {
		go func(ch chan *scipipe.FileTarget) {
			for ft := range ch {
				proc.Out.Chan <- ft
			}
			wg.Done()
		}(inp.Chan)
	}
	wg.Wait()
	proc.Out.Close()
}

func (proc *Merger) Connect(outPort *scipipe.OutPort) {
	inPort := scipipe.NewInPort()
	inPort.Connect(outPort)
	proc.ins = append(proc.ins, inPort)
}
