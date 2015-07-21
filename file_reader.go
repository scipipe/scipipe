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

func (self *fileReader) Run() {
	defer close(self.Out)

	file, err := os.Open(<-self.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		self.Out <- append([]byte(nil), scan.Bytes()...)
	}
	if scan.Err() != nil {
		log.Fatal(scan.Err())
	}
}
