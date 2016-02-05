package scipipe

import (
	"bufio"
	"log"
	"os"
)

// FileReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type FileReader struct {
	Process
	FilePath chan string
	Out      chan []byte
}

// Instantiate a new FileReader
func NewFileReader() *FileReader {
	return &FileReader{
		FilePath: make(chan string),
		Out:      make(chan []byte, BUFSIZE),
	}
}

// Run the FileReader
func (proc *FileReader) Run() {
	defer close(proc.Out)

	file, err := os.Open(<-proc.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		proc.Out <- append([]byte(nil), scan.Bytes()...)
	}
	if scan.Err() != nil {
		log.Fatal(scan.Err())
	}
}
