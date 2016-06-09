package scipipeutils

import (
	"bufio"
	"github.com/scipipe/scipipe"
	"log"
	"os"
)

// FileReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type FileReader struct {
	scipipe.Process
	FilePath chan string
	OutLine  chan []byte
}

// Instantiate a new FileReader
func NewFileReader() *FileReader {
	return &FileReader{
		FilePath: make(chan string),
		OutLine:  make(chan []byte, scipipe.BUFSIZE),
	}
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
