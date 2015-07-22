package scipipe

import (
	"bufio"
	"log"
	"os"
)

type fileReader struct {
	task
	FilePath chan string
	Out      chan []byte
}

func NewFileReader() *fileReader {
	return &fileReader{
		FilePath: make(chan string),
		Out:      make(chan []byte, BUFSIZE),
	}
}

func (proc *fileReader) Run() {
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
