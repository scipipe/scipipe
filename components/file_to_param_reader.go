package components

import (
	"bufio"
	"log"
	"os"

	"github.com/scipipe/scipipe"
)

// FileToParamReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type FileToParamReader struct {
	scipipe.BaseProcess
	filePath string
}

// NewFileToParamReader returns an initialized new FileToParamReader
func NewFileToParamReader(wf *scipipe.Workflow, name string, filePath string) *FileToParamReader {
	p := &FileToParamReader{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		filePath:    filePath,
	}
	p.InitOutParamPort(p, "line")
	wf.AddProc(p)
	return p
}

// OutLine returns an parameter out-port with lines of the files being read
func (p *FileToParamReader) OutLine() *scipipe.OutParamPort { return p.OutParamPort("line") }

// Run the FileToParamReader
func (p *FileToParamReader) Run() {
	defer p.CloseAllOutPorts()

	file, err := os.Open(p.filePath)
	if err != nil {
		err = errWrapf(err, "[FileToParamReader] Could not open file %s", p.filePath)
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		strToSend := scan.Text()
		p.OutLine().Send(strToSend)
	}
	if scan.Err() != nil {
		err = errWrapf(scan.Err(), "[FileToParamReader] Error when scanning input file %s", p.filePath)
		log.Fatal(err)
	}
}
