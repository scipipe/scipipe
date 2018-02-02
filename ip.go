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

// ======= IP ========

// IP (Short for "Information Packet" in Flow-Based Programming terminology)
// contains information and helper methods for a physical file on a normal disk.
type IP struct {
	path      string
	buffer    *bytes.Buffer
	doStream  bool
	lock      *sync.Mutex
	auditInfo *AuditInfo
	SubStream *InPort
}

// NewIP creates a new IP
func NewIP(path string) *IP {
	ip := new(IP)
	ip.path = path
	ip.lock = new(sync.Mutex)
	ip.SubStream = NewInPort("in_substream")
	//Don't init buffer if not needed?
	//buf := make([]byte, 0, 128)
	//ip.buffer = bytes.NewBuffer(buf)
	return ip
}

// GetPath returns the (final) path of the physical file
func (ip *IP) GetPath() string {
	return ip.path
}

// GetTempPath returns the temporary path of the physical file
func (ip *IP) GetTempPath() string {
	return ip.path + ".tmp"
}

// GetFifoPath returns the path to use when a FIFO file is used instead of a
// normal file
func (ip *IP) GetFifoPath() string {
	return ip.path + ".fifo"
}

// GetSize returns the size of an existing file, in bytes
func (ip *IP) GetSize() int64 {
	fi, err := os.Stat(ip.path)
	CheckErr(err)
	return fi.Size()
}

// Open opens the file and returns a file handle (*os.File)
func (ip *IP) Open() *os.File {
	f, err := os.Open(ip.GetPath())
	Check(err, "Could not open file: "+ip.GetPath())
	return f
}

// OpenTemp opens the temp file and returns a file handle (*os.File)
func (ip *IP) OpenTemp() *os.File {
	f, err := os.Open(ip.GetTempPath())
	Check(err, "Could not open temp file: "+ip.GetTempPath())
	return f
}

// OpenWriteTemp opens the file for writing, and returns a file handle (*os.File)
func (ip *IP) OpenWriteTemp() *os.File {
	f, err := os.Create(ip.GetTempPath())
	Check(err, "Could not open temp file for writing: "+ip.GetTempPath())
	return f
}

// Read reads the whole content of the file and returns the content as a byte
// array
func (ip *IP) Read() []byte {
	dat, err := ioutil.ReadFile(ip.GetPath())
	Check(err, "Could not open file for reading: "+ip.GetPath())
	return dat
}

// ReadAuditFile reads the content of the audit file and return it as a byte array
func (ip *IP) ReadAuditFile() []byte {
	dat, err := ioutil.ReadFile(ip.GetAuditFilePath())
	Check(err, "Could not open file for reading: "+ip.GetAuditFilePath())
	return dat
}

// WriteTempFile writes a byte array ([]byte) to the file's temp path
func (ip *IP) WriteTempFile(dat []byte) {
	err := ioutil.WriteFile(ip.GetTempPath(), dat, 0644)
	Check(err, "Could not write to temp file: "+ip.GetTempPath())
}

const (
	sleepDurationSec = 1
)

// Atomize renames the temporary file name to the final file name, thus enabling
// to separate unfinished, and finished files
func (ip *IP) Atomize() {
	Debug.Println("IP: Atomizing", ip.GetTempPath(), "->", ip.GetPath())
	doneAtomizing := false
	for !doneAtomizing {
		if ip.TempFileExists() {
			ip.lock.Lock()
			err := os.Rename(ip.GetTempPath(), ip.path)
			Check(err, "Could not rename file: "+ip.GetTempPath())
			ip.lock.Unlock()
			doneAtomizing = true
			Debug.Println("IP: Done atomizing", ip.GetTempPath(), "->", ip.GetPath())
		} else {
			Debug.Printf("Sleeping for %d seconds before atomizing ...\n", sleepDurationSec)
			time.Sleep(time.Duration(sleepDurationSec) * time.Second)
		}
	}
}

// CreateFifo creates a FIFO file for the IP
func (ip *IP) CreateFifo() {
	ip.lock.Lock()
	cmd := "mkfifo " + ip.GetFifoPath()
	Debug.Println("Now creating FIFO with command:", cmd)

	if _, err := os.Stat(ip.GetFifoPath()); err == nil {
		Warning.Println("FIFO already exists, so not creating a new one:", ip.GetFifoPath())
	} else {
		_, err := exec.Command("bash", "-c", cmd).Output()
		Check(err, "Could not execute command: "+cmd)
	}

	ip.lock.Unlock()
}

// RemoveFifo removes the FIFO file, if it exists
func (ip *IP) RemoveFifo() {
	// FIXME: Shouldn't we check first whether the fifo exists?
	ip.lock.Lock()
	output, err := exec.Command("bash", "-c", "rm "+ip.GetFifoPath()).Output()
	Check(err, "Could not delete fifo file: "+ip.GetFifoPath())
	Debug.Println("Removed FIFO output: ", output)
	ip.lock.Unlock()
}

// Exists checks if the file exists (at its final file name)
func (ip *IP) Exists() bool {
	exists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.GetPath()); err == nil {
		exists = true
	}
	ip.lock.Unlock()
	return exists
}

// TempFileExists checks if the temp-file exists
func (ip *IP) TempFileExists() bool {
	tempFileExists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.GetTempPath()); err == nil {
		tempFileExists = true
	}
	ip.lock.Unlock()
	return tempFileExists
}

// FifoFileExists checks if the FIFO-file (named pipe file) exists
func (ip *IP) FifoFileExists() bool {
	fifoFileExists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.GetFifoPath()); err == nil {
		fifoFileExists = true
	}
	ip.lock.Unlock()
	return fifoFileExists
}

// GetParam returns the parameter named key, from the IPs audit info
func (ip *IP) GetParam(key string) string {
	val, ok := ip.GetAuditInfo().Params[key]
	if !ok {
		Error.Fatalf("Could not find parameter %s in ip with path: %s\n", key, ip.GetPath())
	}
	return val
}

// GetKey returns the key value for the key with key k from the IPs audit info
// (don't confuse this with the keys of maps in go. Keys in this case is a
// SciPipe audit info concept)
func (ip *IP) GetKey(k string) string {
	v, ok := ip.GetAuditInfo().Keys[k]
	if !ok {
		Error.Fatalf("Could not find key %s in ip with path: %s\n", k, ip.GetPath())
	}
	return v
}

// GetKeys returns the audit info's key values
func (ip *IP) GetKeys() map[string]string {
	return ip.GetAuditInfo().Keys
}

// AddKey adds the key k with value v
func (ip *IP) AddKey(k string, v string) {
	ai := ip.GetAuditInfo()
	if ai.Keys[k] != "" && ai.Keys[k] != v {
		Error.Fatalf("Can not add value %s to existing key %s with different value %s\n", v, k, ai.Keys[k])
	}
	ai.Keys[k] = v
}

// AddKeys adds a map of keys to the IPs audit info
func (ip *IP) AddKeys(keys map[string]string) {
	for k, v := range keys {
		ip.AddKey(k, v)
	}
}

// UnMarshalJSON is a helper function to unmarshal the content of the IPs file
// to the interface v
func (ip *IP) UnMarshalJSON(v interface{}) {
	d := ip.Read()
	err := json.Unmarshal(d, v)
	Check(err, "Could not unmarshal content of file: "+ip.GetPath())
}

// GetAuditInfo returns the AuditInfo struct for the IP
func (ip *IP) GetAuditInfo() *AuditInfo {
	defer ip.lock.Unlock()
	ip.lock.Lock()
	if ip.auditInfo == nil {
		ip.auditInfo = NewAuditInfo()
		auditFileData, err := ioutil.ReadFile(ip.GetAuditFilePath())
		if err == nil {
			unmarshalErr := json.Unmarshal(auditFileData, ip.auditInfo)
			Check(unmarshalErr, "Could not unmarshal audit log file content: "+ip.GetAuditFilePath())
		}
	}
	return ip.auditInfo
}

// SetAuditInfo sets the AuditInfo struct for the IP
func (ip *IP) SetAuditInfo(ai *AuditInfo) {
	ip.lock.Lock()
	ip.auditInfo = ai
	ip.lock.Unlock()
}

// GetAuditFilePath returns the file path of the audit info file for the IP
func (ip *IP) GetAuditFilePath() string {
	return ip.GetPath() + ".audit.json"
}

// WriteAuditLogToFile writes the audit log to its designated file
func (ip *IP) WriteAuditLogToFile() {
	auditInfo := ip.GetAuditInfo()
	auditInfoJSON, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
	Check(jsonErr, "Could not marshall JSON")
	writeErr := ioutil.WriteFile(ip.GetAuditFilePath(), auditInfoJSON, 0644)
	Check(writeErr, "Could not write audit file: "+ip.GetPath())
}

// ======= IPGen =======

// IPGen is initialized by a set of strings with file paths, and from that will
// return instantiated (generated) IP on its Out-port, when run.
type IPGen struct {
	name      string
	Out       *OutPort
	FilePaths []string
}

// NewIPGen initializes a new IPGen component from a list of file paths
func NewIPGen(workflow *Workflow, name string, filePaths ...string) (fq *IPGen) {
	fq = &IPGen{
		name:      name,
		Out:       NewOutPort("out"),
		FilePaths: filePaths,
	}
	workflow.AddProc(fq)
	return
}

// Run runs the IPGen process, returning instantiated IP
func (ipg *IPGen) Run() {
	defer ipg.Out.Close()
	for _, fp := range ipg.FilePaths {
		ipg.Out.Send(NewIP(fp))
	}
}

// Name returns the name of the IPGen process
func (ipg *IPGen) Name() string {
	return ipg.name
}

// IsConnected check if the IPGen outport is connected
func (ipg *IPGen) IsConnected() bool {
	return ipg.Out.IsConnected()
}
