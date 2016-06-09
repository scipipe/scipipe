package scipipeutils

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
	In       chan []byte
	FilePath chan string
}

func NewFileWriter() *FileWriter {
	return &FileWriter{
		FilePath: make(chan string),
	}
}

func NewFileWriterFromPath(path string) *FileWriter {
	t := &FileWriter{
		FilePath: make(chan string),
	}
	go func() {
		t.FilePath <- path
	}()
	return t
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
