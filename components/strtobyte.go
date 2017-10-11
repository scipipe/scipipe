package components

import "github.com/scipipe/scipipe"

// SciPipe component that converts packets of string type to byte
type strToByte struct {
	scipipe.Process
	name string
	In   chan string
	Out  chan []byte
}

func (p *strToByte) Name() string {
	return p.name
}

func NewStrToByte(wf *scipipe.Workflow, name string) *strToByte {
	stb := &strToByte{
		name: name,
		In:   make(chan string, scipipe.BUFSIZE),
		Out:  make(chan []byte, scipipe.BUFSIZE),
	}
	wf.AddProc(stb)
	return stb
}

func (proc *strToByte) Run() {
	defer close(proc.Out)
	for line := range proc.In {
		proc.Out <- []byte(line)
	}
}
