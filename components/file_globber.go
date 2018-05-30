package components

import (
	"github.com/scipipe/scipipe"
	"path/filepath"
)

// FileGlobber is initiated with a set of glob patterns paths, which it will
// use to find concrete file paths, for which it will return a stream of
// corresponding File IPs on its outport Out()
type FileGlobber struct {
	scipipe.BaseProcess
	globPatterns []string
}

// NewFileGlobber returns a new initialized FileGlobber process
func NewFileGlobber(wf *scipipe.Workflow, name string, globPatterns ...string) *FileGlobber {
	if len(globPatterns) < 1 {
		scipipe.Failf("FileGlobber with name '%s': No glob paths supplied! Must take at least one glob path. You might also have forgot to provide a name to the fileglobber.", name)
	}
	p := &FileGlobber{
		BaseProcess:  scipipe.NewBaseProcess(wf, name),
		globPatterns: globPatterns,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port, on which file IPs based on the file paths the
// process was initialized with, will be retrieved.
func (p *FileGlobber) Out() *scipipe.OutPort { return p.OutPort("out") }

// Run runs the FileGlobber process
func (p *FileGlobber) Run() {
	defer p.CloseAllOutPorts()
	for _, globPtn := range p.globPatterns {
		scipipe.Audit.Printf("%s: Globbing for files, with pattern: %s", p.Name(), globPtn)
		matches, err := filepath.Glob(globPtn)
		scipipe.CheckWithMsg(err, "FileGlobber: This glob pattern doesn't look right: "+globPtn)
		for _, filePath := range matches {
			scipipe.Audit.Printf("%s: Sending concrete file %s", p.Name(), filePath)
			p.Out().Send(scipipe.NewFileIP(filePath))
		}
	}
}
