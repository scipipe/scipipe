package scipipe

import (
	"bufio"
	"log"
	"os"
)

type fileWriter struct {
	InFilePath chan string
	In         chan []byte
}

func NewFileWriter() *fileWriter {
	return &fileWriter{
		InFilePath: make(chan string),
		In:         make(chan []byte, BUFSIZE),
	}
}

func (self *fileWriter) Run() {
	f, err := os.Create(<-self.InFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for line := range self.In {
		w.Write(line)
	}
	w.Flush()
}
