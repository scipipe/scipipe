package components

import "github.com/scipipe/scipipe"

// FanOut sends *FileTarget received on the InFile in-bound port, and sends
// them on all out-ports created via the Out method.
type FanOut struct {
	scipipe.Process
	InFile   *scipipe.FilePort
	outPorts map[string]*scipipe.FilePort
}

// NewFanOut creates a new FanOut process
func NewFanOut() *FanOut {
	return &FanOut{
		InFile:   scipipe.NewFilePort(),
		outPorts: make(map[string]*scipipe.FilePort),
	}
}

func (p *FanOut) Out(portName string) *scipipe.FilePort {
	if p.outPorts[portName] == nil {
		p.outPorts[portName] = scipipe.NewFilePort()
	}
	return p.outPorts[portName]
}

// Run runs the FanOut process
func (proc *FanOut) Run() {
	for _, outPort := range proc.outPorts {
		defer close(outPort.Chan)
	}

	for ft := range proc.InFile.Chan {
		for key, outPort := range proc.outPorts {
			scipipe.Debug.Println("FanOut: Sending file ", ft.GetPath(), " on out-port ", key)
			outPort.Chan <- ft
		}
	}
}

func (proc *FanOut) IsConnected() bool {
	isConnected := true
	if !proc.InFile.IsConnected() {
		isConnected = false
	} else {
		for portName, port := range proc.outPorts {
			if !port.IsConnected() {
				scipipe.Error.Printf("FanOut: Port %s is not connected!\n", portName)
				isConnected = false
			}
		}
	}
	return isConnected
}
