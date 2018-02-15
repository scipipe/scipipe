package scipipe

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

// ======= IP ========

type IP interface {
	Id() string
	LocalPath() string
	ContentHash() string
	UpdateContentHash()
	ReadyToWrite() bool
	Content() []byte   // Good for params, etc
	MakeReadyToWrite() // Whatever needs to be done before opening/writing to the file
	MakeReadyToRead()
	FixStuffAfterWriting() // Whatever needs to be done after a file is finished writing to
	FixStuffAfterReading() // Whatever needs to be done after a file is finished writing to
}

// LocalFileIP (Short for "Information Packet" in Flow-Based Programming terminology)
// contains information and helper methods for a physical file on a normal disk.
type LocalFileIP struct {
	BaseIP
	path string
}

// NewLocalFileIP creates a new IP, to represent a file on the local file system
func NewLocalFileIP(path string) *LocalFileIP {
	ip := &LocalFileIP{
		BaseIP: NewBaseIP(path),
		path:   path,
	}
	return ip
}

// Path returns the (final) path of the physical file
func (ip *LocalFileIP) Path() string {
	return ip.path
}

// TempPath returns the temporary path of the physical file
func (ip *LocalFileIP) TempPath() string {
	return ip.path + ".tmp"
}

// FifoPath returns the path to use when a FIFO file is used instead of a
// normal file
func (ip *LocalFileIP) FifoPath() string {
	return ip.path + ".fifo"
}

// Size returns the size of an existing file, in bytes
func (ip *LocalFileIP) Size() int64 {
	fi, err := os.Stat(ip.path)
	Check(err)
	return fi.Size()
}

// Open opens the file and returns a file handle (*os.File)
func (ip *LocalFileIP) Open() *os.File {
	f, err := os.Open(ip.Path())
	CheckWithMsg(err, "Could not open file: "+ip.Path())
	return f
}

// OpenTemp opens the temp file and returns a file handle (*os.File)
func (ip *LocalFileIP) OpenTemp() *os.File {
	f, err := os.Open(ip.TempPath())
	CheckWithMsg(err, "Could not open temp file: "+ip.TempPath())
	return f
}

// OpenWriteTemp opens the file for writing, and returns a file handle (*os.File)
func (ip *LocalFileIP) OpenWriteTemp() *os.File {
	f, err := os.Create(ip.TempPath())
	CheckWithMsg(err, "Could not open temp file for writing: "+ip.TempPath())
	return f
}

// Read reads the whole content of the file and returns the content as a byte
// array
func (ip *LocalFileIP) Read() []byte {
	dat, err := ioutil.ReadFile(ip.Path())
	CheckWithMsg(err, "Could not open file for reading: "+ip.Path())
	return dat
}

// ReadAuditFile reads the content of the audit file and return it as a byte array
func (ip *LocalFileIP) ReadAuditFile() []byte {
	dat, err := ioutil.ReadFile(ip.AuditFilePath())
	CheckWithMsg(err, "Could not open file for reading: "+ip.AuditFilePath())
	return dat
}

// WriteTempFile writes a byte array ([]byte) to the file's temp path
func (ip *LocalFileIP) WriteTempFile(dat []byte) {
	err := ioutil.WriteFile(ip.TempPath(), dat, 0644)
	CheckWithMsg(err, "Could not write to temp file: "+ip.TempPath())
}

const (
	sleepDurationSec = 1
)

// Atomize renames the temporary file name to the final file name, thus enabling
// to separate unfinished, and finished files
func (ip *LocalFileIP) Atomize() {
	Debug.Println("IP: Atomizing", ip.TempPath(), "->", ip.Path())
	doneAtomizing := false
	for !doneAtomizing {
		if ip.TempFileExists() {
			ip.lock.Lock()
			err := os.Rename(ip.TempPath(), ip.path)
			CheckWithMsg(err, "Could not rename file: "+ip.TempPath())
			ip.lock.Unlock()
			doneAtomizing = true
			Debug.Println("IP: Done atomizing", ip.TempPath(), "->", ip.Path())
		} else {
			Debug.Printf("Sleeping for %d seconds before atomizing ...\n", sleepDurationSec)
			time.Sleep(time.Duration(sleepDurationSec) * time.Second)
		}
	}
}

// CreateFifo creates a FIFO file for the IP
func (ip *LocalFileIP) CreateFifo() {
	ip.lock.Lock()
	cmd := "mkfifo " + ip.FifoPath()
	Debug.Println("Now creating FIFO with command:", cmd)

	if _, err := os.Stat(ip.FifoPath()); err == nil {
		Warning.Println("FIFO already exists, so not creating a new one:", ip.FifoPath())
	} else {
		_, err := exec.Command("bash", "-c", cmd).Output()
		CheckWithMsg(err, "Could not execute command: "+cmd)
	}

	ip.lock.Unlock()
}

// RemoveFifo removes the FIFO file, if it exists
func (ip *LocalFileIP) RemoveFifo() {
	// FIXME: Shouldn't we check first whether the fifo exists?
	ip.lock.Lock()
	output, err := exec.Command("bash", "-c", "rm "+ip.FifoPath()).Output()
	CheckWithMsg(err, "Could not delete fifo file: "+ip.FifoPath())
	Debug.Println("Removed FIFO output: ", output)
	ip.lock.Unlock()
}

// Exists checks if the file exists (at its final file name)
func (ip *LocalFileIP) Exists() bool {
	exists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.Path()); err == nil {
		exists = true
	}
	ip.lock.Unlock()
	return exists
}

// TempFileExists checks if the temp-file exists
func (ip *LocalFileIP) TempFileExists() bool {
	tempFileExists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.TempPath()); err == nil {
		tempFileExists = true
	}
	ip.lock.Unlock()
	return tempFileExists
}

// FifoFileExists checks if the FIFO-file (named pipe file) exists
func (ip *LocalFileIP) FifoFileExists() bool {
	fifoFileExists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.FifoPath()); err == nil {
		fifoFileExists = true
	}
	ip.lock.Unlock()
	return fifoFileExists
}

// --------------------------------------------------------------------------------
// IPGenerator helper process
// --------------------------------------------------------------------------------

// IPGenerator is initialized by a set of strings with file paths, and from that will
// return instantiated (generated) IP on its Out-port, when run.
type IPGenerator struct {
	BaseProcess
	FilePaths []string
}

// NewIPGenerator initializes a new IPGenerator component from a list of file paths
func NewIPGenerator(wf *Workflow, name string, filePaths ...string) (p *IPGenerator) {
	p = &IPGenerator{
		BaseProcess: NewBaseProcess(wf, name),
		FilePaths:   filePaths,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port of the IPGenerator
func (p *IPGenerator) Out() *OutPort {
	return p.OutPort("out")
}

// Run runs the IPGenerator process, returning instantiated IP
func (p *IPGenerator) Run() {
	defer p.Out().Close()
	for _, fp := range p.FilePaths {
		p.Out().Send(NewLocalFileIP(fp))
	}
}

// UnMarshalJSON is a helper function to unmarshal the content of the IPs file
// to the interface v
func (ip *LocalFileIP) UnMarshalJSON(v interface{}) {
	d := ip.Read()
	err := json.Unmarshal(d, v)
	CheckWithMsg(err, "Could not unmarshal content"+ip.Path())
}
