package components

import (
	"bufio"
	"log"
	"os"

	"github.com/scipipe/scipipe"
)

// FileToParamsReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type FileToParamsReader struct {
	scipipe.BaseProcess
	filePath string
}

// NewFileToParamsReader returns an initialized new FileToParamsReader
func NewFileToParamsReader(wf *scipipe.Workflow, name string, filePath string) *FileToParamsReader {
	p := &FileToParamsReader{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		filePath:    filePath,
	}
	p.InitOutParamPort(p, "line")
	wf.AddProc(p)
	return p
}

// OutLine returns an parameter out-port with lines of the files being read
func (p *FileToParamsReader) OutLine() *scipipe.OutParamPort { return p.OutParamPort("line") }

// Run the FileToParamsReader
func (p *FileToParamsReader) Run() {
	defer p.CloseAllOutPorts()

	file, err := os.Open(p.filePath)
	if err != nil {
		err = errWrapf(err, "[FileToParamsReader] Could not open file %s", p.filePath)
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		strToSend := scan.Text()
		p.OutLine().Send(strToSend)
	}
	if scan.Err() != nil {
		err = errWrapf(scan.Err(), "[FileToParamsReader] Error when scanning input file %s", p.filePath)
		log.Fatal(err)
	}
}
