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

// InPort returns the in-port with name portName
func (p *BaseProcess) InPort(portName string) *InPort {
	if p.inPorts[portName] == nil {
		Failf("No such in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.inPorts[portName]
}

// InitInPort adds the in-port port to the process, with name portName
func (p *BaseProcess) InitInPort(proc WorkflowProcess, portName string) {
	if p.inPorts[portName] != nil {
		Failf("Such an in-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	ipt := NewInPort(portName)
	ipt.process = proc
	p.inPorts[portName] = ipt
}

// InPorts returns a map of all the in-ports of the process, keyed by their
// names
func (p *BaseProcess) InPorts() map[string]*InPort {
	return p.inPorts
}

// DeleteInPort deletes an InPort object from the process
func (p *BaseProcess) DeleteInPort(portName string) {
	if _, ok := p.inPorts[portName]; !ok {
		Failf("No such in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	delete(p.inPorts, portName)
}

// ------------------------------------------------
// Out-port stuff
// ------------------------------------------------

// InitOutPort adds the out-port port to the process, with name portName
func (p *BaseProcess) InitOutPort(proc WorkflowProcess, portName string) {
	if _, ok := p.outPorts[portName]; ok {
		Failf("Such an out-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	opt := NewOutPort(portName)
	opt.process = proc
	p.outPorts[portName] = opt
}

// OutPort returns the out-port with name portName
func (p *BaseProcess) OutPort(portName string) *OutPort {
	if _, ok := p.outPorts[portName]; !ok {
		Failf("No such out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.outPorts[portName]
}

// OutPorts returns a map of all the out-ports of the process, keyed by their
// names
func (p *BaseProcess) OutPorts() map[string]*OutPort {
	return p.outPorts
}

// DeleteOutPort deletes a OutPort object from the process
func (p *BaseProcess) DeleteOutPort(portName string) {
	if _, ok := p.outPorts[portName]; !ok {
		Failf("No such out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	delete(p.outPorts, portName)
}

// ------------------------------------------------
// Param-in-port stuff
// ------------------------------------------------

// InitParamInPort adds the parameter port paramPort with name portName
func (p *BaseProcess) InitParamInPort(proc WorkflowProcess, portName string) {
	if _, ok := p.paramInPorts[portName]; ok {
		Failf("Such a param-in-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	pip := NewParamInPort(portName)
	pip.process = proc
	p.paramInPorts[portName] = pip
}

// ParamInPort returns the parameter port with name portName
func (p *BaseProcess) ParamInPort(portName string) *ParamInPort {
	if _, ok := p.paramInPorts[portName]; !ok {
		Failf("No such param-in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.paramInPorts[portName]
}

// ParamInPorts returns all parameter in-ports of the process
func (p *BaseProcess) ParamInPorts() map[string]*ParamInPort {
	return p.paramInPorts
}

// DeleteParamInPort deletes a ParamInPort object from the process
func (p *BaseProcess) DeleteParamInPort(portName string) {
	if _, ok := p.paramInPorts[portName]; !ok {
		Failf("No such param-in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	delete(p.paramInPorts, portName)
}

// ------------------------------------------------
// Param-out-port stuff
// ------------------------------------------------

// InitParamOutPort initializes the parameter port paramPort with name portName
// to the process We need to supply the concrete process used here as well,
// since this method might be used as part of an embedded struct, meaning that
// the process in the receiver is just the *BaseProcess, which doesn't suffice.
func (p *BaseProcess) InitParamOutPort(proc WorkflowProcess, portName string) {
	if _, ok := p.paramOutPorts[portName]; ok {
		Failf("Such a param-out-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	pop := NewParamOutPort(portName)
	pop.process = proc
	p.paramOutPorts[portName] = pop
}

// ParamOutPort returns the parameter port with name portName
func (p *BaseProcess) ParamOutPort(portName string) *ParamOutPort {
	if _, ok := p.paramOutPorts[portName]; !ok {
		Failf("No such param-out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.paramOutPorts[portName]
}

// ParamOutPorts returns all parameter out-ports of the process
func (p *BaseProcess) ParamOutPorts() map[string]*ParamOutPort {
	return p.paramOutPorts
}

// DeleteParamOutPort deletes a ParamOutPort object from the process
func (p *BaseProcess) DeleteParamOutPort(portName string) {
	if _, ok := p.paramOutPorts[portName]; !ok {
		Failf("No such param-out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	delete(p.paramOutPorts, portName)
}

// ------------------------------------------------
// Other stuff
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

// CloseOutPorts closes all (normal) out-ports
func (p *BaseProcess) CloseOutPorts() {
	for _, p := range p.OutPorts() {
		p.Close()
	}
}

// CloseParamOutPorts closes all parameter out-ports
func (p *BaseProcess) CloseParamOutPorts() {
	for _, op := range p.ParamOutPorts() {
		op.Close()
	}
}

// CloseAllOutPorts closes all normal-, and parameter out ports
func (p *BaseProcess) CloseAllOutPorts() {
	p.CloseOutPorts()
	p.CloseParamOutPorts()
}

func (p *BaseProcess) receiveOnInPorts() (ips map[string]*FileIP, inPortsOpen bool) {
	inPortsOpen = true
	ips = make(map[string]*FileIP)
	// Read input IPs on in-ports and set up path mappings
	for inpName, inPort := range p.InPorts() {
		Debug.Printf("Process %s: Receieving on inPort %s ...", p.name, inpName)
		ip, open := <-inPort.Chan
		if !open {
			inPortsOpen = false
			continue
		}
		Debug.Printf("Process %s: Got ip %s ...", p.name, ip.Path())
		ips[inpName] = ip
	}
	return
}

func (p *BaseProcess) receiveOnParamInPorts() (params map[string]string, paramPortsOpen bool) {
	paramPortsOpen = true
	params = make(map[string]string)
	// Read input IPs on in-ports and set up path mappings
	for pname, pport := range p.ParamInPorts() {
		pval, open := <-pport.Chan
		if !open {
			paramPortsOpen = false
			continue
		}
		Debug.Printf("Process %s: Got param %s ...", p.name, pval)
		params[pname] = pval
	}
	return
}
