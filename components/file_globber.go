package components

import (
	"path/filepath"

	"github.com/scipipe/scipipe"
)

type FileGlobber struct {
	scipipe.Process
	Out         *scipipe.FilePort
	globPattern string
}

func NewFileGlobber(globPattern string) *FileGlobber {
	return &FileGlobber{
		Out:         scipipe.NewFilePort(),
		globPattern: globPattern,
	}
}

func (p *FileGlobber) Run() {
	defer p.Out.Close()

	matches, err := filepath.Glob(p.globPattern)
	scipipe.CheckErr(err)

	for _, m := range matches {
		ip := scipipe.NewInformationPacket(m)
		p.Out.Chan <- ip
	}
}

func (p *FileGlobber) IsConnected() bool {
	return p.Out.IsConnected()
}
