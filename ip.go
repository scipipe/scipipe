package scipipe

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// IP Is the base interface which all other IPs need to adhere to
type IP interface {
	ID() string
	Atomize()
	// ----------------------------------------
	// Some tentative additions:
	// ----------------------------------------
	// Digest() string
	// URL() string        // Example: file:///proj/cheminf/exp/20180101-logd/dat/rawdata.tsv
	// StagedPath() string // Example: /proj/cheminf/exp/20180101-logd/dat/rawdata.tsv
	// TempPath() string
	// EnsureStaged()
	// EnsureUnstaged()
	// ParamType() int // string / int8 / float64 / bool / date? / time?
	// Param() string
	// Key() string
	// Keys() map[string]string
	// AddKey(key string)
	// AddKeys(keys ...string)
	// AuditInfo()
	// SetAuditInfo()
	// AuditInfoFilePath() string
	// WriteAuditLogToFile()
	// Params() map[string]string
	// ----------------------------------------
	// Persistable() bool // Whether the content or data of the IP can be persisted, e.g. by being written to disk
	// Read() []byte
	// Write(data []byte)
	// OpenR() io.Reader // Return a reader interface to read content directly
	// OpenW() io.Writer // (Possibly relevant for object storage without staging)
	// OpenRW() io.ReadWriter
	// Close()
	// ----------------------------------------
}

// ------------------------------------------------------------------------
// BaseIP type
// ------------------------------------------------------------------------

// BaseIP contains foundational functionality which all IPs need to implement.
// It is meant to be embedded into other IP implementations.
type BaseIP struct {
	path      string
	id        string
	auditInfo *AuditInfo
}

// NewBaseIP creates a new BaseIP
func NewBaseIP(path string) *BaseIP {
	return &BaseIP{
		path: path,
		id:   randSeqLC(20),
	}
}

// ID returns a globally unique ID for the IP
func (ip *BaseIP) ID() string {
	return ip.id
}

// ------------------------------------------------------------------------
// FileIP type
// ------------------------------------------------------------------------

// FileIP (Short for "Information Packet" in Flow-Based Programming terminology)
// contains information and helper methods for a physical file on a normal disk.
type FileIP struct {
	*BaseIP
	buffer    *bytes.Buffer
	doStream  bool
	lock      *sync.Mutex
	SubStream *InPort
}

// NewFileIP creates a new FileIP
func NewFileIP(path string) *FileIP {
	ip := &FileIP{
		BaseIP:    NewBaseIP(path),
		lock:      &sync.Mutex{},
		SubStream: NewInPort("in_substream"),
	}
	if ip.Exists() {
		ip.AuditInfo() // This will populate the audit info from file
	}
	//Don't init buffer if not needed?
	//buf := make([]byte, 0, 128)
	//ip.buffer = bytes.NewBuffer(buf)
	return ip
}

// ------------------------------------------------------------------------
// Path stuff
// ------------------------------------------------------------------------

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

// ------------------------------------------------------------------------
// Check-thing stuff
// ------------------------------------------------------------------------

// Size returns the size of an existing file, in bytes
func (ip *FileIP) Size() int64 {
	fi, err := os.Stat(ip.path)
	Check(err)
	return fi.Size()
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

// ------------------------------------------------------------------------
// Open file-stuff
// ------------------------------------------------------------------------

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
	ip.createDirs()
	f, err := os.Create(ip.TempPath())
	CheckWithMsg(err, "Could not open temp file for writing: "+ip.TempPath())
	return f
}

// ------------------------------------------------------------------------
// FIFO-specific stuff
// ------------------------------------------------------------------------

// CreateFifo creates a FIFO file for the FileIP
func (ip *FileIP) CreateFifo() {
	ip.createDirs()
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

// ------------------------------------------------------------------------
// Read/Write stuff
// ------------------------------------------------------------------------

// Read reads the whole content of the file and returns the content as a byte
// array
func (ip *FileIP) Read() []byte {
	dat, err := ioutil.ReadFile(ip.Path())
	CheckWithMsg(err, "Could not open file for reading: "+ip.Path())
	return dat
}

// Write writes a byte array ([]byte) to the file's temp file path
func (ip *FileIP) Write(dat []byte) {
	ip.createDirs()
	err := ioutil.WriteFile(ip.TempPath(), dat, 0644)
	CheckWithMsg(err, "Could not write to temp file: "+ip.TempPath())
}

const (
	maxTries      = 3
	backoffFactor = 4
)

// Atomize renames the temporary file name to the final file name, thus enabling
// to separate unfinished, and finished files
func (ip *FileIP) Atomize() {
	Debug.Println("FileIP: Atomizing", ip.TempPath(), "->", ip.Path())
	doneAtomizing := false
	tries := 0

	sleepDurationSec := 1
	for !doneAtomizing {
		if ip.TempFileExists() {
			ip.lock.Lock()
			err := os.Rename(ip.TempPath(), ip.path)
			CheckWithMsg(err, "Could not rename file: "+ip.TempPath())
			ip.lock.Unlock()
			doneAtomizing = true
			Debug.Println("FileIP: Done atomizing", ip.TempPath(), "->", ip.Path())
		} else {
			if tries >= maxTries {
				Failf("Failed to find .tmp file after %d tries, so shutting down: %s\nNote: If this problem persists, it could be a problem with your workflow, that the configured output filename in scipipe doesn't match what is written by the tool.\n", maxTries, ip.TempPath())
			}
			Warning.Printf("Expected .tmp file missing: %s\nSleeping for %d seconds before checking again ...\n", ip.TempPath(), sleepDurationSec)
			time.Sleep(time.Duration(sleepDurationSec) * time.Second)
			sleepDurationSec *= backoffFactor
			tries += 1
		}
	}
}

// ------------------------------------------------------------------------
// Params and keys
// ------------------------------------------------------------------------

// Param returns the parameter named key, from the IPs audit info
func (ip *FileIP) Param(key string) string {
	val, ok := ip.AuditInfo().Params[key]
	if !ok {
		Failf("Could not find parameter %s in ip with path: %s\n", key, ip.Path())
	}
	return val
}

// ------------------------------------------------------------------------
// Keys stuff
// ------------------------------------------------------------------------

// Key returns the key value for the key with key k from the IPs audit info
// (don't confuse this with the keys of maps in go. Keys in this case is a
// SciPipe audit info concept)
func (ip *FileIP) Key(k string) string {
	v, ok := ip.AuditInfo().Keys[k]
	if !ok {
		Failf("Could not find key %s in ip with path: %s\n", k, ip.Path())
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
		Failf("Can not add value %s to existing key %s with different value %s\n", v, k, ai.Keys[k])
	}
	ai.Keys[k] = v
}

// AddKeys adds a map of keys to the IPs audit info
func (ip *FileIP) AddKeys(keys map[string]string) {
	for k, v := range keys {
		ip.AddKey(k, v)
	}
}

// ------------------------------------------------------------------------
// AuditInfo stuff
// ------------------------------------------------------------------------

// AuditFilePath returns the file path of the audit info file for the FileIP
func (ip *FileIP) AuditFilePath() string {
	return ip.Path() + ".audit.json"
}

// SetAuditInfo sets the AuditInfo struct for the FileIP
func (ip *FileIP) SetAuditInfo(ai *AuditInfo) {
	ip.lock.Lock()
	ip.auditInfo = ai
	ip.lock.Unlock()
}

// WriteAuditLogToFile writes the audit log to its designated file
func (ip *FileIP) WriteAuditLogToFile() {
	auditInfo := ip.AuditInfo()
	auditInfoJSON, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
	CheckWithMsg(jsonErr, "Could not marshall JSON")
	ip.createDirs()
	writeErr := ioutil.WriteFile(ip.AuditFilePath(), auditInfoJSON, 0644)
	CheckWithMsg(writeErr, "Could not write audit file: "+ip.Path())
}

// AuditInfo returns the AuditInfo struct for the FileIP
func (ip *FileIP) AuditInfo() *AuditInfo {
	defer ip.lock.Unlock()
	ip.lock.Lock()
	if ip.auditInfo == nil {
		ip.auditInfo = NewAuditInfo()
		auditFileData, readFileErr := ioutil.ReadFile(ip.AuditFilePath())
		if readFileErr != nil {
			if os.IsNotExist(readFileErr) {
				Info.Printf("Audit file not found, so not unmarshalling: %s\n", ip.AuditFilePath())
			} else {
				Failf("Could not read audit file, which does exist: %s", ip.AuditFilePath())
			}
		} else {
			unmarshalErr := json.Unmarshal(auditFileData, ip.auditInfo)
			CheckWithMsg(unmarshalErr, "Could not unmarshal audit log file content: "+ip.AuditFilePath())
		}
	}
	return ip.auditInfo
}

// ------------------------------------------------------------------------
// Extra convenience functions
// ------------------------------------------------------------------------

// UnMarshalJSON is a helper function to unmarshal the content of the IPs file
// to the interface v
func (ip *FileIP) UnMarshalJSON(v interface{}) {
	d := ip.Read()
	err := json.Unmarshal(d, v)
	CheckWithMsg(err, "Could not unmarshal content of file: "+ip.Path())
}

// ------------------------------------------------------------------------
// Helper functions
// ------------------------------------------------------------------------

// CreateDirs creates all directories needed to enable writing the IP to its
// path (or temporary-path, which will have the same directory)
func (ip *FileIP) createDirs() {
	dir := filepath.Dir(ip.Path())
	err := os.MkdirAll(dir, 0777)
	CheckWithMsg(err, "Could not create directory: "+dir)
}
