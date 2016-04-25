package scipipe

import (
	"sync"
)

type Merger struct {
	Ins [](chan *FileTarget)
	Out chan *FileTarget
}

func NewMerger() *Merger {
	return &Merger{
		Out: make(chan *FileTarget, BUFSIZE),
	}
}

func (proc *Merger) Run() {
	var wg sync.WaitGroup
	wg.Add(len(proc.Ins))
	for _, ch := range proc.Ins {
		go func(ch chan *FileTarget) {
			for ft := range ch {
				proc.Out <- ft
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(proc.Out)
}
