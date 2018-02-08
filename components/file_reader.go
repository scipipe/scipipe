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
	scipipe.BaseProcess
}

// NewFileReader returns an initialized new FileReader
func NewFileReader(wf *scipipe.Workflow, name string) *FileReader {
	p := &FileReader{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
	}
	p.InitParamInPort(p, "filepath")
	p.InitParamOutPort(p, "line")
	wf.AddProc(p)
	return p
}

// InFilePath returns the parameter in-port on which a file name is read
func (p *FileReader) InFilePath() *scipipe.ParamInPort { return p.ParamInPort("filepath") }

// OutLine returns an parameter out-port with lines of the files being read
func (p *FileReader) OutLine() *scipipe.ParamOutPort { return p.ParamOutPort("line") }

// Run the FileReader
func (p *FileReader) Run() {
	defer p.CloseAllOutPorts()

	file, err := os.Open(<-p.InFilePath().Chan) // Read a single file name right now
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		p.OutLine().Send(scan.Text() + "\n")
	}
	if scan.Err() != nil {
		log.Fatal(scan.Err())
	}
}
