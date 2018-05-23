package components

import (
	"github.com/scipipe/scipipe"
)

// FileSource is initiated with a set of file paths, which it will send as a
// stream of File IPs on its outport Out()
type FileSource struct {
	scipipe.BaseProcess
	filePaths []string
}

// NewFileSource returns a new initialized FileSource process
func NewFileSource(wf *scipipe.Workflow, name string, filePaths ...string) *FileSource {
	p := &FileSource{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		filePaths:   filePaths,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port, on which file IPs based on the file paths the
// process was initialized with, will be retrieved.
func (p *FileSource) Out() *scipipe.OutPort { return p.OutPort("out") }

// Run runs the FileSource process
func (p *FileSource) Run() {
	defer p.CloseAllOutPorts()
	for _, filePath := range p.filePaths {
		p.Out().Send(scipipe.NewFileIP(filePath))
	}
}
