package components

import (
	"path/filepath"

	"github.com/scipipe/scipipe"
)

type FileGlobber struct {
	name        string
	Out         *scipipe.OutPort
	globPattern string
}

func (p *FileGlobber) Name() string {
	return p.name
}

func NewFileGlobber(wf *scipipe.Workflow, name string, globPattern string) *FileGlobber {
	fg := &FileGlobber{
		name:        name,
		Out:         scipipe.NewOutPort(),
		globPattern: globPattern,
	}
	wf.AddProc(fg)
	return fg
}

func (p *FileGlobber) Run() {
	defer p.Out.Close()

	matches, err := filepath.Glob(p.globPattern)
	scipipe.CheckErr(err)

	for _, m := range matches {
		ip := scipipe.NewIP(m)
		p.Out.Send(ip)
	}
}

func (p *FileGlobber) IsConnected() bool {
	return p.Out.IsConnected()
}
