package components

import (
	"bufio"
	"log"
	"os"

	"github.com/scipipe/scipipe"
)

// FileReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type FileReader struct {
	scipipe.WorkflowProcess
	name     string
	FilePath *scipipe.ParamInPort
	OutLine  *scipipe.ParamOutPort
}

// NewFileReader returns an initialized new FileReader
func NewFileReader(wf *scipipe.Workflow, name string) *FileReader {
	fr := &FileReader{
		name:     name,
		FilePath: scipipe.NewParamInPort("filepath"),
		OutLine:  scipipe.NewParamOutPort("line"),
	}
	fr.FilePath.SetProcess(fr)
	fr.OutLine.SetProcess(fr)
	wf.AddProc(fr)
	return fr
}

// Name returns the name of the FileReader process
func (p *FileReader) Name() string {
	return p.name
}

// ParamInPorts returns all parameter in-ports for the process
func (p *FileReader) ParamInPorts() map[string]*scipipe.ParamInPort {
	return map[string]*scipipe.ParamInPort{
		p.FilePath.Name(): p.FilePath,
	}
}

// ParamOutPorts returns all parameter out-ports for the process
func (p *FileReader) ParamOutPorts() map[string]*scipipe.ParamOutPort {
	return map[string]*scipipe.ParamOutPort{
		p.OutLine.Name(): p.OutLine,
	}
}

// Run the FileReader
func (p *FileReader) Run() {
	defer p.OutLine.Close()

	file, err := os.Open(<-p.FilePath.Chan) // Read a single file name right now
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		p.OutLine.Send(scan.Text() + "\n")
	}
	if scan.Err() != nil {
		log.Fatal(scan.Err())
	}
}
