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
	FilePath chan string
	OutLine  chan []byte
}

// Instantiate a new FileReader
func NewFileReader(wf *scipipe.Workflow, name string) *FileReader {
	fr := &FileReader{
		name:     name,
		FilePath: make(chan string),
		OutLine:  make(chan []byte, scipipe.BUFSIZE),
	}
	wf.AddProc(fr)
	return fr
}

func (proc *FileReader) Name() string {
	return proc.name
}

// Run the FileReader
func (proc *FileReader) Run() {
	defer close(proc.OutLine)

	file, err := os.Open(<-proc.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		proc.OutLine <- append(append([]byte(nil), scan.Bytes()...), []byte("\n")...)
	}
	if scan.Err() != nil {
		log.Fatal(scan.Err())
	}
}
