package scipipe

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
)

// ======= FileTarget ========

// FileTarget contains information and helper methods for a physical file on a
// normal disk.
type FileTarget struct {
	path     string
	buffer   *bytes.Buffer
	doStream bool
	lock     *sync.Mutex
}

// Create new FileTarget "object"
func NewFileTarget(path string) *FileTarget {
	ft := new(FileTarget)
	ft.path = path
	ft.lock = new(sync.Mutex)
	//Don't init buffer if not needed?
	//buf := make([]byte, 0, 128)
	//ft.buffer = bytes.NewBuffer(buf)
	return ft
}

// Get the (final) path of the physical file
func (ft *FileTarget) GetPath() string {
	return ft.path
}

// Get the temporary path of the physical file
func (ft *FileTarget) GetTempPath() string {
	return ft.path + ".tmp"
}

// Get the path to use when a FIFO file is used instead of a normal file
func (ft *FileTarget) GetFifoPath() string {
	return ft.path + ".fifo"
}

// Open the file and return a file handle (*os.File)
func (ft *FileTarget) Open() *os.File {
	f, err := os.Open(ft.GetPath())
	Check(err)
	return f
}

// Open the temp file and return a file handle (*os.File)
func (ft *FileTarget) OpenTemp() *os.File {
	f, err := os.Open(ft.GetTempPath())
	Check(err)
	return f
}

// Open the file for writing return a file handle (*os.File)
func (ft *FileTarget) OpenWriteTemp() *os.File {
	f, err := os.Create(ft.GetTempPath())
	Check(err)
	return f
}

// Read the whole content of the file and return as a byte array ([]byte)
func (ft *FileTarget) Read() []byte {
	dat, err := ioutil.ReadFile(ft.GetPath())
	Check(err)
	return dat
}

// Write a byte array ([]byte) to the file (first to its temp path, and then atomize)
func (ft *FileTarget) WriteTempFile(dat []byte) {
	err := ioutil.WriteFile(ft.GetTempPath(), dat, 0644)
	Check(err)
}

// Change from the temporary file name to the final file name
func (ft *FileTarget) Atomize() {
	Debug.Println("FileTarget: Atomizing", ft.GetTempPath(), "->", ft.GetPath())
	ft.lock.Lock()
	err := os.Rename(ft.GetTempPath(), ft.path)
	Check(err)
	ft.lock.Unlock()
	Debug.Println("FileTarget: Done atomizing", ft.GetTempPath(), "->", ft.GetPath())
}

// Create FIFO file for the FileTarget
func (ft *FileTarget) CreateFifo() {
	ft.lock.Lock()
	cmd := "mkfifo " + ft.GetFifoPath()
	Debug.Println("Now creating FIFO with command:", cmd)

	if _, err := os.Stat(ft.GetFifoPath()); err == nil {
		Warning.Println("FIFO already exists, so not creating a new one:", ft.GetFifoPath())
	} else {
		_, err := exec.Command("bash", "-c", cmd).Output()
		Check(err)
	}

	ft.lock.Unlock()
}

// Remove the FIFO file, if it exists
func (ft *FileTarget) RemoveFifo() {
	// FIXME: Shouldn't we check first whether the fifo exists?
	ft.lock.Lock()
	output, err := exec.Command("bash", "-c", "rm "+ft.GetFifoPath()).Output()
	Check(err)
	Debug.Println("Removed FIFO output: ", output)
	ft.lock.Unlock()
}

// Check if the file exists (at its final file name)
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

// FileQueue is initialized by a set of strings with file paths, and from that
// will return instantiated FileTargets on its Out-port, when run.
type FileQueue struct {
	Process
	Out       *OutPort
	FilePaths []string
}

// Initialize a new FileQueue component from a list of file paths
func NewFileQueue(filePaths ...string) (fq *FileQueue) {
	fq = &FileQueue{
		Out:       NewOutPort(),
		FilePaths: filePaths,
	}
	return
}

// Execute the FileQueue, returning instantiated FileTargets
func (proc *FileQueue) Run() {
	defer proc.Out.Close()
	for _, fp := range proc.FilePaths {
		proc.Out.Chan <- NewFileTarget(fp)
	}
}

// Check if the fileQueue outport is connected
func (proc *FileQueue) IsConnected() bool {
	return proc.Out.IsConnected()
}

func (proc *Sink) deleteInPortAtKey(i int) {
	proc.inPorts = append(proc.inPorts[:i], proc.inPorts[i+1:]...)
}
