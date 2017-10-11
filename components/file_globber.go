package components

import (
	"path/filepath"

	"github.com/scipipe/scipipe"
)

type FileGlobber struct {
	scipipe.Process
	name        string
	Out         *scipipe.FilePort
	globPattern string
}

func (p *FileGlobber) Name() string {
	return p.name
}

func NewFileGlobber(wf *scipipe.Workflow, name string, globPattern string) *FileGlobber {
	fg := &FileGlobber{
		name:        name,
		Out:         scipipe.NewFilePort(),
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
		ip := scipipe.NewInformationPacket(m)
		p.Out.Send(ip)
	}
}

func (p *FileGlobber) IsConnected() bool {
	return p.Out.IsConnected()
}
