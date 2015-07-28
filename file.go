package scipipe

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"
)

// ======= FileTarget ========

type FileTarget struct {
	path   string
	buffer *bytes.Buffer
	stream bool
}

func NewFileTarget(path string) *FileTarget {
	ft := new(FileTarget)
	ft.path = path
	//Don't init buffer if not needed?
	//buf := make([]byte, 0, 128)
	//ft.buffer = bytes.NewBuffer(buf)
	return ft
}

func (ft *FileTarget) GetPath() string {
	return ft.path
}

func (ft *FileTarget) GetTempPath() string {
	return ft.path + ".tmp"
}

func (ft *FileTarget) Open() *os.File {
	f, err := os.Open(ft.GetPath())
	Check(err)
	return f
}

func (ft *FileTarget) Read() []byte {
	dat, err := ioutil.ReadFile(ft.GetPath())
	Check(err)
	return dat
}

func (ft *FileTarget) Write(dat []byte) {
	err := ioutil.WriteFile(ft.GetTempPath(), dat, 0644)
	ft.Atomize()
	Check(err)
}

func (ft *FileTarget) Atomize() {
	time.Sleep(0 * time.Millisecond) // TODO: Remove in production. Just for demo purposes!
	err := os.Rename(ft.GetTempPath(), ft.path)
	Check(err)
}

func (ft *FileTarget) Exists() bool {
	if _, err := os.Stat(ft.GetPath()); err == nil {
		return true
	}
	return false
}

// ======= FileQueue =======

type FileQueue struct {
	process
	Out       chan *FileTarget
	FilePaths []string
}

func FQ(fps ...string) (fq *FileQueue) {
	return NewFileQueue(fps...)
}

func NewFileQueue(fps ...string) (fq *FileQueue) {
	filePaths := []string{}
	for _, fp := range fps {
		filePaths = append(filePaths, fp)
	}
	fq = &FileQueue{
		Out:       make(chan *FileTarget, BUFSIZE),
		FilePaths: filePaths,
	}
	return
}

func (proc *FileQueue) Run() {
	defer close(proc.Out)
	for _, fp := range proc.FilePaths {
		proc.Out <- NewFileTarget(fp)
	}
}

// ======= Sink =======

type Sink struct {
	process
	In chan *FileTarget
}

func NewSink() (s *Sink) {
	return &Sink{}
}

func (proc *Sink) Run() {
	for ft := range proc.In {
		Debug.Println("Received file in sink: ", ft.GetPath())
	}
}
