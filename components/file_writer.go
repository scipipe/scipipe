package components

import (
	"bufio"
	"fmt"
	"github.com/scipipe/scipipe"
	"log"
	"os"
)

// FileWriter takes a file path on its FilePath in-port, file contents on its In in-port
// and write the file contents to a file with the specified path.
type FileWriter struct {
	scipipe.Process
	name     string
	In       chan []byte
	FilePath chan string
}

func (p *FileWriter) Name() string {
	return p.name
}

func NewFileWriter(wf *scipipe.Workflow, name string) *FileWriter {
	fw := &FileWriter{
		name:     name,
		FilePath: make(chan string),
	}
	wf.AddProc(fw)
	return fw
}

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

func (proc *FileWriter) Run() {
	f, err := os.Create(<-proc.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for line := range proc.In {
		w.WriteString(fmt.Sprint(string(line), "\n"))
	}
	w.Flush()
}

func (proc *FileWriter) IsConnected() bool { return true }
