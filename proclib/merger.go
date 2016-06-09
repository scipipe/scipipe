package proclib

import (
	"github.com/scipipe/scipipe"
	"sync"
)

type Merger struct {
	Ins [](*scipipe.InPort)
	Out *scipipe.OutPort
}

func NewMerger() *Merger {
	return &Merger{}
}

func (proc *Merger) Run() {
	var wg sync.WaitGroup
	wg.Add(len(proc.Ins))
	for _, inp := range proc.Ins {
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
