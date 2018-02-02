package components

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/scipipe/scipipe"
)

// FileWriter takes a file path on its FilePath in-port, file contents on its In in-port
// and write the file contents to a file with the specified path.
type FileWriter struct {
	name     string
	In       chan []byte
	FilePath chan string
}

// Name returns the name of the FileWriter process
func (p *FileWriter) Name() string {
	return p.name
}

// NewFileWriter returns an initialized FileWriter process, which will read file
// IPs on its in-port In, and write those to files, for which the paths are
// retrieved on the FilePath channel
func NewFileWriter(wf *scipipe.Workflow, name string) *FileWriter {
	fw := &FileWriter{
		name:     name,
		FilePath: make(chan string),
	}
	wf.AddProc(fw)
	return fw
}

// NewFileWriterFromPath will create a new FileWriter process, and send to it a
// single file name, path
func NewFileWriterFromPath(wf *scipipe.Workflow, path string) *FileWriter {
	fw := &FileWriter{
		FilePath: make(chan string),
	}
	go func() {
		fw.FilePath <- path
	}()
	wf.AddProc(fw)
	return fw
}

// Run runs the FileWriter process
func (p *FileWriter) Run() {
	f, err := os.Create(<-p.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for line := range p.In {
		w.WriteString(fmt.Sprint(string(line), "\n"))
	}
	w.Flush()
}

// IsConnected tells whether all the ports of the FileWriter process are all connected
func (p *FileWriter) IsConnected() bool { return true }
