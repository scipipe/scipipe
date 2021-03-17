package scipipe

// BaseProcess provides a skeleton for processes, such as the main Process
// component, and the custom components in the scipipe/components library
type BaseProcess struct {
	name          string
	workflow      *Workflow
	inPorts       map[string]*InPort
	outPorts      map[string]*OutPort
	inParamPorts  map[string]*InParamPort
	outParamPorts map[string]*OutParamPort
}

// NewBaseProcess returns a new BaseProcess, connected to the provided workflow,
// and with the name name
func NewBaseProcess(wf *Workflow, name string) BaseProcess {
	return BaseProcess{
		workflow:      wf,
		name:          name,
		inPorts:       make(map[string]*InPort),
		outPorts:      make(map[string]*OutPort),
		inParamPorts:  make(map[string]*InParamPort),
		outParamPorts: make(map[string]*OutParamPort),
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
	if _, ok := p.inPorts[portName]; !ok {
		p.failf("No such in-port ('%s'). Please check your workflow code!\n", portName)
	}
	return p.inPorts[portName]
}

// InitInPort adds the in-port port to the process, with name portName
func (p *BaseProcess) InitInPort(proc WorkflowProcess, portName string) {
	if _, ok := p.inPorts[portName]; ok {
		p.failf("Such an in-port ('%s') already exists. Please check your workflow code!\n", portName)
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
		p.failf("No such in-port ('%s'). Please check your workflow code!\n", portName)
	}
	delete(p.inPorts, portName)
}

// ------------------------------------------------
// Out-port stuff
// ------------------------------------------------

// InitOutPort adds the out-port port to the process, with name portName
func (p *BaseProcess) InitOutPort(proc WorkflowProcess, portName string) {
	if _, ok := p.outPorts[portName]; ok {
		p.failf("Such an out-port ('%s') already exists. Please check your workflow code!\n", portName)
	}
	opt := NewOutPort(portName)
	opt.process = proc
	p.outPorts[portName] = opt
}

// OutPort returns the out-port with name portName
func (p *BaseProcess) OutPort(portName string) *OutPort {
	if _, ok := p.outPorts[portName]; !ok {
		p.failf("No such out-port ('%s'). Please check your workflow code!\n", portName)
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
		p.failf("No such out-port ('%s'). Please check your workflow code!\n", portName)
	}
	delete(p.outPorts, portName)
}

// ------------------------------------------------
// Param-in-port stuff
// ------------------------------------------------

// InitInParamPort adds the parameter port paramPort with name portName
func (p *BaseProcess) InitInParamPort(proc WorkflowProcess, portName string) {
	if _, ok := p.inParamPorts[portName]; ok {
		p.failf("Such a param-in-port ('%s') already exists. Please check your workflow code!\n", portName)
	}
	pip := NewInParamPort(portName)
	pip.process = proc
	p.inParamPorts[portName] = pip
}

// InParamPort returns the parameter port with name portName
func (p *BaseProcess) InParamPort(portName string) *InParamPort {
	if _, ok := p.inParamPorts[portName]; !ok {
		p.failf("No such param-in-port ('%s'). Please check your workflow code!\n", portName)
	}
	return p.inParamPorts[portName]
}

// InParamPorts returns all parameter in-ports of the process
func (p *BaseProcess) InParamPorts() map[string]*InParamPort {
	return p.inParamPorts
}

// DeleteInParamPort deletes a InParamPort object from the process
func (p *BaseProcess) DeleteInParamPort(portName string) {
	if _, ok := p.inParamPorts[portName]; !ok {
		p.failf("No such param-in-port ('%s'). Please check your workflow code!\n", portName)
	}
	delete(p.inParamPorts, portName)
}

// ------------------------------------------------
// Param-out-port stuff
// ------------------------------------------------

// InitOutParamPort initializes the parameter port paramPort with name portName
// to the process We need to supply the concrete process used here as well,
// since this method might be used as part of an embedded struct, meaning that
// the process in the receiver is just the *BaseProcess, which doesn't suffice.
func (p *BaseProcess) InitOutParamPort(proc WorkflowProcess, portName string) {
	if _, ok := p.outParamPorts[portName]; ok {
		p.failf("Such a param-out-port ('%s') already exists. Please check your workflow code!\n", portName)
	}
	pop := NewOutParamPort(portName)
	pop.process = proc
	p.outParamPorts[portName] = pop
}

// OutParamPort returns the parameter port with name portName
func (p *BaseProcess) OutParamPort(portName string) *OutParamPort {
	if _, ok := p.outParamPorts[portName]; !ok {
		p.failf("No such param-out-port ('%s'). Please check your workflow code!\n", portName)
	}
	return p.outParamPorts[portName]
}

// OutParamPorts returns all parameter out-ports of the process
func (p *BaseProcess) OutParamPorts() map[string]*OutParamPort {
	return p.outParamPorts
}

// DeleteOutParamPort deletes a OutParamPort object from the process
func (p *BaseProcess) DeleteOutParamPort(portName string) {
	if _, ok := p.outParamPorts[portName]; !ok {
		p.failf("No such param-out-port ('%s'). Please check your workflow code!\n", portName)
	}
	delete(p.outParamPorts, portName)
}

// ------------------------------------------------
// Other stuff
// ------------------------------------------------

// Ready checks whether all the process' ports are connected
func (p *BaseProcess) Ready() (isReady bool) {
	isReady = true
	for portName, port := range p.inPorts {
		if !port.Ready() {
			Error.Printf("InPort %s of process %s is not connected - check your workflow code!\n", portName)
			isReady = false
		}
	}
	for portName, port := range p.outPorts {
		if !port.Ready() {
			Error.Printf("OutPort %s of process %s is not connected - check your workflow code!\n", portName)
			isReady = false
		}
	}
	for portName, port := range p.inParamPorts {
		if !port.Ready() {
			Error.Printf("InParamPort %s of process %s is not connected - check your workflow code!\n", portName)
			isReady = false
		}
	}
	for portName, port := range p.outParamPorts {
		if !port.Ready() {
			Error.Printf("OutParamPort %s of process %s is not connected - check your workflow code!\n", portName)
			isReady = false
		}
	}
	return isReady
}

// CloseOutPorts closes all (normal) out-ports
func (p *BaseProcess) CloseOutPorts() {
	for _, p := range p.OutPorts() {
		p.Close()
	}
}

// CloseOutParamPorts closes all parameter out-ports
func (p *BaseProcess) CloseOutParamPorts() {
	for _, op := range p.OutParamPorts() {
		op.Close()
	}
}

// CloseAllOutPorts closes all normal-, and parameter out ports
func (p *BaseProcess) CloseAllOutPorts() {
	p.CloseOutPorts()
	p.CloseOutParamPorts()
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

func (p *BaseProcess) receiveOnInParamPorts() (params map[string]string, paramPortsOpen bool) {
	paramPortsOpen = true
	params = make(map[string]string)
	// Read input IPs on in-ports and set up path mappings
	for pname, pport := range p.InParamPorts() {
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

// failf Prints an error message with the process as context, before exiting
// the program
func (p *BaseProcess) failf(errMsg string, strs ...string) {
	Failf("[Process:%s]: "+errMsg, append([]string{p.name}, strs...))
}
