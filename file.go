package scipipe

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
)

// ======= FileTarget ========

type FileTarget struct {
	path     string
	buffer   *bytes.Buffer
	doStream bool
	lock     *sync.Mutex
}

func NewFileTarget(path string) *FileTarget {
	ft := new(FileTarget)
	ft.path = path
	ft.lock = new(sync.Mutex)
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

func (ft *FileTarget) GetFifoPath() string {
	return ft.path + ".fifo"
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
	Debug.Println("FileTarget: Atomizing", ft.GetTempPath(), "->", ft.GetPath())
	ft.lock.Lock()
	err := os.Rename(ft.GetTempPath(), ft.path)
	Check(err)
	ft.lock.Unlock()
	Debug.Println("FileTarget: Done atomizing", ft.GetTempPath(), "->", ft.GetPath())
}

func (ft *FileTarget) CreateFifo() {
	ft.lock.Lock()
	cmd := "mkfifo " + ft.GetFifoPath()
	Debug.Println("Now creating FIFO with command:", cmd)

	if _, err := os.Stat(ft.GetFifoPath()); err == nil {
		Warn.Println("FIFO already exists, so cannot be created:", ft.GetFifoPath())
	} else {
		_, err := exec.Command("bash", "-c", cmd).Output()
		Check(err)
	}

	ft.lock.Unlock()
}

func (ft *FileTarget) RemoveFifo() {
	ft.lock.Lock()
	output, err := exec.Command("bash", "-c", "rm "+ft.GetFifoPath()).Output()
	Check(err)
	Debug.Println("Removed FIFO output: ", output)
	ft.lock.Unlock()
}

func (ft *FileTarget) Exists() bool {
	exists := false
	ft.lock.Lock()
	if _, err := os.Stat(ft.GetPath()); err == nil {
		exists = true
	}
	ft.lock.Unlock()
	return exists
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
