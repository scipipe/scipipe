package scipipe

// FanOut sends *FileTarget received on the InFile in-bound port, and sends
// them on all out-ports created via the GetOutPort method.
type FanOut struct {
	Process
	InFile   *InPort
	outPorts map[string]*OutPort
}

// NewFanOut creates a new FanOut process
func NewFanOut() *FanOut {
	return &FanOut{
		InFile:   NewInPort(),
		outPorts: make(map[string]*OutPort),
	}
}

func (p *FanOut) GetOutPort(portName string) *OutPort {
	if p.outPorts[portName] == nil {
		p.outPorts[portName] = NewOutPort()
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
			Debug.Println("FanOut: Sending file ", ft.GetPath(), " on out-port ", key)
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
				Error.Printf("FanOut: Port %s is not connected!", portName)
				isConnected = false
			}
		}
	}
	return isConnected
}
