package scipipe

// BaseProcess provides a skeleton for processes, such as the main Process
// component, and the custom components in the scipipe/components library
type BaseProcess struct {
	name          string
	workflow      *Workflow
	inPorts       map[string]*InPort
	outPorts      map[string]*OutPort
	paramInPorts  map[string]*ParamInPort
	paramOutPorts map[string]*ParamOutPort
}

// NewBaseProcess returns a new BaseProcess, connected to the provided workflow,
// and with the name name
func NewBaseProcess(wf *Workflow, name string) BaseProcess {
	return BaseProcess{
		workflow:      wf,
		name:          name,
		inPorts:       make(map[string]*InPort),
		outPorts:      make(map[string]*OutPort),
		paramInPorts:  make(map[string]*ParamInPort),
		paramOutPorts: make(map[string]*ParamOutPort),
	}
}

// Name returns the name of the process
func (p *BaseProcess) Name() string {
	return p.name
}

// Workflow returns the workflow the process is connected to
func (p *BaseProcess) Workflow() *Workflow {
	return p.workflow
}

// ------------------------------------------------
// In-port stuff
// ------------------------------------------------

// In returns the in-port with name portName
func (p *BaseProcess) In(portName string) *InPort {
	if p.inPorts[portName] == nil {
		Error.Fatalf("No such in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.inPorts[portName]
}

// SetInPort adds the in-port port to the process, with name portName
func (p *BaseProcess) SetInPort(portName string, port *InPort) {
	if p.inPorts[portName] != nil {
		Error.Fatalf("Such an in-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	p.inPorts[portName] = port
}

// InPorts returns a map of all the in-ports of the process, keyed by their
// names
func (p *BaseProcess) InPorts() map[string]*InPort {
	return p.inPorts
}

// ------------------------------------------------
// Out-port stuff
// ------------------------------------------------

// Out returns the out-port with name portName
func (p *BaseProcess) Out(portName string) *OutPort {
	if p.outPorts[portName] == nil {
		Error.Fatalf("No such out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.outPorts[portName]
}

// SetOutPort adds the out-port port to the process, with name portName
func (p *BaseProcess) SetOutPort(portName string, port *OutPort) {
	if p.outPorts[portName] != nil {
		Error.Fatalf("Such an out-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	p.outPorts[portName] = port
}

// OutPorts returns a map of all the out-ports of the process, keyed by their
// names
func (p *BaseProcess) OutPorts() map[string]*OutPort {
	return p.outPorts
}

// ------------------------------------------------
// Param-in-port stuff
// ------------------------------------------------

// ParamInPort returns the parameter port with name paramPortName
func (p *BaseProcess) ParamInPort(paramPortName string) *ParamInPort {
	if p.paramInPorts[paramPortName] == nil {
		Error.Fatalf("No such param-port ('%s') for process '%s'. Please check your workflow code!\n", paramPortName, p.name)
	}
	return p.paramInPorts[paramPortName]
}

// ParamInPorts returns all parameter ports of the process
func (p *BaseProcess) ParamInPorts() map[string]*ParamInPort {
	return p.paramInPorts
}

// SetParamInPort adds the parameter port paramPort with name paramPortName
func (p *BaseProcess) SetParamInPort(paramPortName string, paramPort *ParamInPort) {
	p.paramInPorts[paramPortName] = paramPort
}

// ------------------------------------------------
// Param-out-port stuff
// ------------------------------------------------

// ParamOutPorts returns an empty map of ParamOutPorts, to comlpy with the
// WorkflowProcess interface (since param-out-ports are not applicable for
// normal processes)
func (p *BaseProcess) ParamOutPorts() map[string]*ParamOutPort {
	return p.paramOutPorts
}

// ------------------------------------------------
// Sanity check stuff
// ------------------------------------------------

// Connected checks whether all the process' ports are connected
func (p *BaseProcess) Connected() (isConnected bool) {
	isConnected = true
	for portName, port := range p.inPorts {
		if !port.Connected() {
			Error.Printf("InPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	for portName, port := range p.outPorts {
		if !port.Connected() {
			Error.Printf("OutPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	for portName, port := range p.paramInPorts {
		if !port.Connected() {
			Error.Printf("ParamInPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	for portName, port := range p.paramOutPorts {
		if !port.Connected() {
			Error.Printf("ParamOutPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	return isConnected
}
