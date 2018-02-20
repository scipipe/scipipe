package scipipe

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"time"
)

// ======= FileIP ========

// FileIP (Short for "Information Packet" in Flow-Based Programming terminology)
// contains information and helper methods for a physical file on a normal disk.
type FileIP struct {
	path      string
	buffer    *bytes.Buffer
	doStream  bool
	lock      *sync.Mutex
	auditInfo *AuditInfo
	SubStream *InPort
}

// NewFileIP creates a new FileIP
func NewFileIP(path string) *FileIP {
	ip := &FileIP{
		path:      path,
		lock:      &sync.Mutex{},
		SubStream: NewInPort("in_substream"),
	}
	//Don't init buffer if not needed?
	//buf := make([]byte, 0, 128)
	//ip.buffer = bytes.NewBuffer(buf)
	return ip
}

// Path returns the (final) path of the physical file
func (ip *FileIP) Path() string {
	return ip.path
}

// TempPath returns the temporary path of the physical file
func (ip *FileIP) TempPath() string {
	return ip.path + ".tmp"
}

// FifoPath returns the path to use when a FIFO file is used instead of a
// normal file
func (ip *FileIP) FifoPath() string {
	return ip.path + ".fifo"
}

// Size returns the size of an existing file, in bytes
func (ip *FileIP) Size() int64 {
	fi, err := os.Stat(ip.path)
	Check(err)
	return fi.Size()
}

// Open opens the file and returns a file handle (*os.File)
func (ip *FileIP) Open() *os.File {
	f, err := os.Open(ip.Path())
	CheckWithMsg(err, "Could not open file: "+ip.Path())
	return f
}

// OpenTemp opens the temp file and returns a file handle (*os.File)
func (ip *FileIP) OpenTemp() *os.File {
	f, err := os.Open(ip.TempPath())
	CheckWithMsg(err, "Could not open temp file: "+ip.TempPath())
	return f
}

// OpenWriteTemp opens the file for writing, and returns a file handle (*os.File)
func (ip *FileIP) OpenWriteTemp() *os.File {
	f, err := os.Create(ip.TempPath())
	CheckWithMsg(err, "Could not open temp file for writing: "+ip.TempPath())
	return f
}

// Read reads the whole content of the file and returns the content as a byte
// array
func (ip *FileIP) Read() []byte {
	dat, err := ioutil.ReadFile(ip.Path())
	CheckWithMsg(err, "Could not open file for reading: "+ip.Path())
	return dat
}

// ReadAuditFile reads the content of the audit file and return it as a byte array
func (ip *FileIP) ReadAuditFile() []byte {
	dat, err := ioutil.ReadFile(ip.AuditFilePath())
	CheckWithMsg(err, "Could not open file for reading: "+ip.AuditFilePath())
	return dat
}

// WriteTempFile writes a byte array ([]byte) to the file's temp path
func (ip *FileIP) WriteTempFile(dat []byte) {
	err := ioutil.WriteFile(ip.TempPath(), dat, 0644)
	CheckWithMsg(err, "Could not write to temp file: "+ip.TempPath())
}

const (
	sleepDurationSec = 1
)

// Atomize renames the temporary file name to the final file name, thus enabling
// to separate unfinished, and finished files
func (ip *FileIP) Atomize() {
	Debug.Println("FileIP: Atomizing", ip.TempPath(), "->", ip.Path())
	doneAtomizing := false
	for !doneAtomizing {
		if ip.TempFileExists() {
			ip.lock.Lock()
			err := os.Rename(ip.TempPath(), ip.path)
			CheckWithMsg(err, "Could not rename file: "+ip.TempPath())
			ip.lock.Unlock()
			doneAtomizing = true
			Debug.Println("FileIP: Done atomizing", ip.TempPath(), "->", ip.Path())
		} else {
			Debug.Printf("Sleeping for %d seconds before atomizing ...\n", sleepDurationSec)
			time.Sleep(time.Duration(sleepDurationSec) * time.Second)
		}
	}
}

// CreateFifo creates a FIFO file for the FileIP
func (ip *FileIP) CreateFifo() {
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
func (ip *FileIP) RemoveFifo() {
	// FIXME: Shouldn't we check first whether the fifo exists?
	ip.lock.Lock()
	output, err := exec.Command("bash", "-c", "rm "+ip.FifoPath()).Output()
	CheckWithMsg(err, "Could not delete fifo file: "+ip.FifoPath())
	Debug.Println("Removed FIFO output: ", output)
	ip.lock.Unlock()
}

// Exists checks if the file exists (at its final file name)
func (ip *FileIP) Exists() bool {
	exists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.Path()); err == nil {
		exists = true
	}
	ip.lock.Unlock()
	return exists
}

// TempFileExists checks if the temp-file exists
func (ip *FileIP) TempFileExists() bool {
	tempFileExists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.TempPath()); err == nil {
		tempFileExists = true
	}
	ip.lock.Unlock()
	return tempFileExists
}

// FifoFileExists checks if the FIFO-file (named pipe file) exists
func (ip *FileIP) FifoFileExists() bool {
	fifoFileExists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.FifoPath()); err == nil {
		fifoFileExists = true
	}
	ip.lock.Unlock()
	return fifoFileExists
}

// Param returns the parameter named key, from the IPs audit info
func (ip *FileIP) Param(key string) string {
	val, ok := ip.AuditInfo().Params[key]
	if !ok {
		Error.Fatalf("Could not find parameter %s in ip with path: %s\n", key, ip.Path())
	}
	return val
}

// Key returns the key value for the key with key k from the IPs audit info
// (don't confuse this with the keys of maps in go. Keys in this case is a
// SciPipe audit info concept)
func (ip *FileIP) Key(k string) string {
	v, ok := ip.AuditInfo().Keys[k]
	if !ok {
		Error.Fatalf("Could not find key %s in ip with path: %s\n", k, ip.Path())
	}
	return v
}

// Keys returns the audit info's key values
func (ip *FileIP) Keys() map[string]string {
	return ip.AuditInfo().Keys
}

// AddKey adds the key k with value v
func (ip *FileIP) AddKey(k string, v string) {
	ai := ip.AuditInfo()
	if ai.Keys[k] != "" && ai.Keys[k] != v {
		Error.Fatalf("Can not add value %s to existing key %s with different value %s\n", v, k, ai.Keys[k])
	}
	ai.Keys[k] = v
}

// AddKeys adds a map of keys to the IPs audit info
func (ip *FileIP) AddKeys(keys map[string]string) {
	for k, v := range keys {
		ip.AddKey(k, v)
	}
}

// UnMarshalJSON is a helper function to unmarshal the content of the IPs file
// to the interface v
func (ip *FileIP) UnMarshalJSON(v interface{}) {
	d := ip.Read()
	err := json.Unmarshal(d, v)
	CheckWithMsg(err, "Could not unmarshal content of file: "+ip.Path())
}

// AuditInfo returns the AuditInfo struct for the FileIP
func (ip *FileIP) AuditInfo() *AuditInfo {
	defer ip.lock.Unlock()
	ip.lock.Lock()
	if ip.auditInfo == nil {
		ip.auditInfo = NewAuditInfo()
		auditFileData, err := ioutil.ReadFile(ip.AuditFilePath())
		if err == nil {
			unmarshalErr := json.Unmarshal(auditFileData, ip.auditInfo)
			CheckWithMsg(unmarshalErr, "Could not unmarshal audit log file content: "+ip.AuditFilePath())
		}
	}
	return ip.auditInfo
}

// SetAuditInfo sets the AuditInfo struct for the FileIP
func (ip *FileIP) SetAuditInfo(ai *AuditInfo) {
	ip.lock.Lock()
	ip.auditInfo = ai
	ip.lock.Unlock()
}

// AuditFilePath returns the file path of the audit info file for the FileIP
func (ip *FileIP) AuditFilePath() string {
	return ip.Path() + ".audit.json"
}

// WriteAuditLogToFile writes the audit log to its designated file
func (ip *FileIP) WriteAuditLogToFile() {
	auditInfo := ip.AuditInfo()
	auditInfoJSON, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
	CheckWithMsg(jsonErr, "Could not marshall JSON")
	writeErr := ioutil.WriteFile(ip.AuditFilePath(), auditInfoJSON, 0644)
	CheckWithMsg(writeErr, "Could not write audit file: "+ip.Path())
}

// --------------------------------------------------------------------------------
// FileIPGenerator helper process
// --------------------------------------------------------------------------------

// FileIPGenerator is initialized by a set of strings with file paths, and from that will
// return instantiated (generated) FileIP on its Out-port, when run.
type FileIPGenerator struct {
	BaseProcess
	FilePaths []string
}

// NewFileIPGenerator initializes a new FileIPGenerator component from a list of file paths
func NewFileIPGenerator(wf *Workflow, name string, filePaths ...string) (p *FileIPGenerator) {
	p = &FileIPGenerator{
		BaseProcess: NewBaseProcess(wf, name),
		FilePaths:   filePaths,
	}
	p.InitOutPort(p, "out")
	wf.AddProc(p)
	return p
}

// Out returns the out-port of the FileIPGenerator
func (p *FileIPGenerator) Out() *OutPort {
	return p.OutPort("out")
}

// Run runs the FileIPGenerator process, returning instantiated FileIP
func (p *FileIPGenerator) Run() {
	defer p.Out().Close()
	for _, fp := range p.FilePaths {
		p.Out().Send(NewFileIP(fp))
	}
}
