package components

import (
	"path/filepath"

	"github.com/scipipe/scipipe"
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

// NewFileGlobberDependent returns a new FileGlobber that depends on upstream
// files to be received on the InPort InDependency() before it starts globbing files.
func NewFileGlobberDependent(wf *scipipe.Workflow, name string, globPatterns ...string) *FileGlobber {
	p := NewFileGlobber(wf, name, globPatterns...)
	p.InitInPort(p, "in_dep")
	return p
}

// Out returns the out-port, on which file IPs based on the file paths the
// process was initialized with, will be retrieved.
func (p *FileGlobber) Out() *scipipe.OutPort { return p.OutPort("out") }

// InDependency takes files which it will wait for before it starts to execute.
func (p *FileGlobber) InDependency() *scipipe.InPort { return p.InPort("in_dep") }

// Run runs the FileGlobber process
func (p *FileGlobber) Run() {
	defer p.CloseAllOutPorts()
	// If we have an InDependency in-port, then loop on the in-channel of that, to make
	// the process wait for IPs on that.
	if _, ok := p.InPorts()["in_dep"]; ok {
		for range p.InDependency().Chan {
			// Do nothing, just empty the channel
		}
	}
	p.globFiles()
}

func (p *FileGlobber) globFiles() {
	for _, globPtn := range p.globPatterns {
		p.Auditf("Globbing for files, with pattern: %s", globPtn)
		matches, err := filepath.Glob(globPtn)
		scipipe.CheckWithMsg(err, "FileGlobber: This glob pattern doesn't look right: "+globPtn)
		for _, filePath := range matches {
			p.Auditf("Sending concrete file %s", filePath)
			newIP, err := scipipe.NewFileIP(filePath)
			if err != nil {
				p.Fail(err)
			}
			p.Out().Send(newIP)
		}
	}
}
