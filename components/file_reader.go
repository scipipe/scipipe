package components

import (
	"bufio"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/scipipe/scipipe"
)

// FileReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type FileReader struct {
	scipipe.BaseProcess
	filePath string
}

// NewFileReader returns an initialized new FileReader
func NewFileReader(wf *scipipe.Workflow, name string, filePath string) *FileReader {
	p := &FileReader{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		filePath:    filePath,
	}
	p.InitInParamPort(p, "filepath")
	p.InitOutParamPort(p, "line")
	wf.AddProc(p)
	return p
}

// OutLine returns an parameter out-port with lines of the files being read
func (p *FileReader) OutLine() *scipipe.OutParamPort { return p.OutParamPort("line") }

// Run the FileReader
func (p *FileReader) Run() {
	defer p.CloseAllOutPorts()

	file, err := os.Open(p.filePath)
	if err != nil {
		err = errors.Wrapf(err, "[FileReader] Could not open file %s", p.filePath)
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		p.OutLine().Send(scan.Text() + "\n")
	}
	if scan.Err() != nil {
		err = errors.Wrapf(scan.Err(), "[FileReader] Error when scanning input file %s", p.filePath)
		log.Fatal(err)
	}
}
