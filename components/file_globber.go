package components

import (
	"path/filepath"

	"github.com/scipipe/scipipe"
)

// FileGlobber is a process that returns multiple files as IPs on its out-port
// Out, based on the glob pattern in globPattern, which is set through the
// factory method NewFileGlobber
type FileGlobber struct {
	name        string
	Out         *scipipe.OutPort
	globPattern string
}

// Name returns the name of the FileGlobber process
func (p *FileGlobber) Name() string {
	return p.name
}

// NewFileGlobber returns a new, initialized FileGlobber process, that will
// return files on its out-port Out, based on the glob pattern in globPattern
func NewFileGlobber(wf *scipipe.Workflow, name string, globPattern string) *FileGlobber {
	fg := &FileGlobber{
		name:        name,
		Out:         scipipe.NewOutPort("out"),
		globPattern: globPattern,
	}
	wf.AddProc(fg)
	return fg
}

// Run runs the FileGlobber process
func (p *FileGlobber) Run() {
	defer p.Out.Close()

	matches, err := filepath.Glob(p.globPattern)
	scipipe.CheckErr(err)

	for _, m := range matches {
		ip := scipipe.NewIP(m)
		p.Out.Send(ip)
	}
}

// IsConnected tells whether all ports of the FileGlobber process are connected
func (p *FileGlobber) IsConnected() bool {
	return p.Out.IsConnected()
}
