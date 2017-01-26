package scipipe

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
)

// ======= InformationPacket ========

// InformationPacket contains information and helper methods for a physical file on a
// normal disk.
type InformationPacket struct {
	path      string
	buffer    *bytes.Buffer
	doStream  bool
	lock      *sync.Mutex
	auditInfo *AuditInfo
}

// Create new InformationPacket "object"
func NewInformationPacket(path string) *InformationPacket {
	ip := new(InformationPacket)
	ip.path = path
	ip.lock = new(sync.Mutex)
	//Don't init buffer if not needed?
	//buf := make([]byte, 0, 128)
	//ip.buffer = bytes.NewBuffer(buf)
	return ip
}

// Get the (final) path of the physical file
func (ip *InformationPacket) GetPath() string {
	return ip.path
}

// Get the temporary path of the physical file
func (ip *InformationPacket) GetTempPath() string {
	return ip.path + ".tmp"
}

// Get the path to use when a FIFO file is used instead of a normal file
func (ip *InformationPacket) GetFifoPath() string {
	return ip.path + ".fifo"
}

// Open the file and return a file handle (*os.File)
func (ip *InformationPacket) Open() *os.File {
	f, err := os.Open(ip.GetPath())
	Check(err, "Could not open file: "+ip.GetPath())
	return f
}

// Open the temp file and return a file handle (*os.File)
func (ip *InformationPacket) OpenTemp() *os.File {
	f, err := os.Open(ip.GetTempPath())
	Check(err, "Could not open temp file: "+ip.GetTempPath())
	return f
}

// Open the file for writing return a file handle (*os.File)
func (ip *InformationPacket) OpenWriteTemp() *os.File {
	f, err := os.Create(ip.GetTempPath())
	Check(err, "Could not open temp file for writing: "+ip.GetTempPath())
	return f
}

// Read the whole content of the file and return as a byte array ([]byte)
func (ip *InformationPacket) Read() []byte {
	dat, err := ioutil.ReadFile(ip.GetPath())
	Check(err, "Could not open file for reading: "+ip.GetPath())
	return dat
}

// Write a byte array ([]byte) to the file (first to its temp path, and then atomize)
func (ip *InformationPacket) WriteTempFile(dat []byte) {
	err := ioutil.WriteFile(ip.GetTempPath(), dat, 0644)
	Check(err, "Could not write to temp file: "+ip.GetTempPath())
}

// Change from the temporary file name to the final file name
func (ip *InformationPacket) Atomize() {
	Debug.Println("InformationPacket: Atomizing", ip.GetTempPath(), "->", ip.GetPath())
	ip.lock.Lock()
	err := os.Rename(ip.GetTempPath(), ip.path)
	Check(err, "Could not rename file: "+ip.GetTempPath())
	ip.lock.Unlock()
	Debug.Println("InformationPacket: Done atomizing", ip.GetTempPath(), "->", ip.GetPath())
}

// Create FIFO file for the InformationPacket
func (ip *InformationPacket) CreateFifo() {
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

// Remove the FIFO file, if it exists
func (ip *InformationPacket) RemoveFifo() {
	// FIXME: Shouldn't we check first whether the fifo exists?
	ip.lock.Lock()
	output, err := exec.Command("bash", "-c", "rm "+ip.GetFifoPath()).Output()
	Check(err, "Could not delete fifo file: "+ip.GetFifoPath())
	Debug.Println("Removed FIFO output: ", output)
	ip.lock.Unlock()
}

// Check if the file exists (at its final file name)
func (ip *InformationPacket) Exists() bool {
	exists := false
	ip.lock.Lock()
	if _, err := os.Stat(ip.GetPath()); err == nil {
		exists = true
	}
	ip.lock.Unlock()
	return exists
}

func (ip *InformationPacket) GetAuditInfo() *AuditInfo {
	return ip.auditInfo
}

func (ip *InformationPacket) SetAuditInfo(ai *AuditInfo) {
	ip.auditInfo = ai
}

func (ip *InformationPacket) WriteAuditLogToFile() {
	auditInfo := ip.GetAuditInfo()
	auditInfoJson, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
	Check(jsonErr, "Could not marshall JSON")
	writeErr := ioutil.WriteFile(ip.GetPath()+".audit.json", auditInfoJson, 0644)
	Check(writeErr, "Could not write audit file: "+ip.GetPath())
}

// ======= IPQueue =======

// IPQueue is initialized by a set of strings with file paths, and from that
// will return instantiated InformationPacket on its Out-port, when run.
type IPQueue struct {
	Process
	Out       *FilePort
	FilePaths []string
}

// Initialize a new IPQueue component from a list of file paths
func NewIPQueue(filePaths ...string) (fq *IPQueue) {
	fq = &IPQueue{
		Out:       NewFilePort(),
		FilePaths: filePaths,
	}
	return
}

// Execute the IPQueue, returning instantiated InformationPacket
func (ipq *IPQueue) Run() {
	defer ipq.Out.Close()
	for _, fp := range ipq.FilePaths {
		ipq.Out.Chan <- NewInformationPacket(fp)
	}
}

// Check if the IPQueue outport is connected
func (ipq *IPQueue) IsConnected() bool {
	return ipq.Out.IsConnected()
}

func (snk *Sink) deleteInPortAtKey(i int) {
	if snk.inPorts != nil {
		if snk.inPorts[i] != nil {
			snk.inPorts = append(snk.inPorts[:i], snk.inPorts[i+1:]...)
		} else {
			Warning.Println("Inport %d does not exist, in sink")
		}
	} else {
		Warning.Println("Inports array not initialized!")
	}
}
