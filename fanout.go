package scipipe

// FanOut sends *FileTarget received on the InFile in-bound port, and sends
// them on all out-ports created via the GetOutPort method.
type FanOut struct {
	Process
	InFile   chan *FileTarget
	outPorts map[string]chan *FileTarget
}

// NewFanOut creates a new FanOut process
func NewFanOut() *FanOut {
	return &FanOut{
		InFile:   make(chan *FileTarget, BUFSIZE),
		outPorts: make(map[string]chan *FileTarget, BUFSIZE),
	}
}

func (p *FanOut) GetOutPort(portName string) chan *FileTarget {
	if p.outPorts[portName] == nil {
		p.outPorts[portName] = make(chan *FileTarget, BUFSIZE)
	}
	return p.outPorts[portName]
}

// Run runs the FanOut process
func (proc *FanOut) Run() {
	for _, outPort := range proc.outPorts {
		defer close(outPort)
	}

	for ft := range proc.InFile {
		for key, outPort := range proc.outPorts {
			Debug.Println("FanOut: Sending file ", ft.GetPath(), " on out-port ", key)
			outPort <- ft
		}
	}
}
