package scipipe

import (
	"sync"
)

// ------------------------------------------------------------------------
// InPort
// ------------------------------------------------------------------------

// InPort represents a pluggable connection to multiple out-ports from other
// processes, from its own process, and with which it is communicating via
// channels under the hood
type InPort struct {
	Chan        chan *FileIP
	name        string
	process     WorkflowProcess
	RemotePorts map[string]*OutPort
	connected   bool
	closeLock   sync.Mutex
}

// NewInPort returns a new InPort struct
func NewInPort(name string) *InPort {
	inp := &InPort{
		name:        name,
		RemotePorts: map[string]*OutPort{},
		Chan:        make(chan *FileIP, BUFSIZE), // This one will contain merged inputs from inChans
		connected:   false,
	}
	return inp
}

// Name returns the name of the InPort
func (pt *InPort) Name() string {
	return pt.Process().Name() + "." + pt.name
}

// Process returns the process connected to the port
func (pt *InPort) Process() WorkflowProcess {
	if pt.process == nil {
		Failf("In-port %s has no connected process", pt.name)
	}
	return pt.process
}

// SetProcess sets the process of the port to p
func (pt *InPort) SetProcess(p WorkflowProcess) {
	pt.process = p
}

// AddRemotePort adds a remote OutPort to the InPort
func (pt *InPort) AddRemotePort(rpt *OutPort) {
	if pt.RemotePorts[rpt.Name()] != nil {
		Failf("[Process:%s]: A remote port with name %s already exists, for param-port %s connected to process %s\n", pt.Process().Name(), rpt.Name(), pt.Name(), rpt.Process().Name())
	}
	pt.RemotePorts[rpt.Name()] = rpt
}

// Connect connects an OutPort to the InPort
func (pt *InPort) Connect(rpt *OutPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetConnectedStatus(true)
	rpt.SetConnectedStatus(true)
}

// Disconnect disconnects the (out-)port with name rptName, from the InPort
func (pt *InPort) Disconnect(rptName string) {
	pt.removeRemotePort(rptName)
	if len(pt.RemotePorts) == 0 {
		pt.SetConnectedStatus(false)
	}
}

// removeRemotePort removes the (out-)port with name rptName, from the InPort
func (pt *InPort) removeRemotePort(rptName string) {
	delete(pt.RemotePorts, rptName)
}

// SetConnectedStatus sets the connected status of the InPort
func (pt *InPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

// Connected tells whether the port is connected or not
func (pt *InPort) Connected() bool {
	return pt.connected
}

// Send sends IPs to the in-port, and is supposed to be called from the remote
// (out-) port, to send to this in-port
func (pt *InPort) Send(ip *FileIP) {
	pt.Chan <- ip
}

// Recv receives IPs from the port
func (pt *InPort) Recv() *FileIP {
	return <-pt.Chan
}

// CloseConnection closes the connection to the remote out-port with name
// rptName, on the InPort
func (pt *InPort) CloseConnection(rptName string) {
	pt.closeLock.Lock()
	delete(pt.RemotePorts, rptName)
	if len(pt.RemotePorts) == 0 {
		close(pt.Chan)
	}
	pt.closeLock.Unlock()
}

// ------------------------------------------------------------------------
// OutPort
// ------------------------------------------------------------------------

// OutPort represents a pluggable connection to multiple in-ports from other
// processes, from its own process, and with which it is communicating via
// channels under the hood
type OutPort struct {
	name        string
	process     WorkflowProcess
	RemotePorts map[string]*InPort
	connected   bool
}

// NewOutPort returns a new OutPort struct
func NewOutPort(name string) *OutPort {
	outp := &OutPort{
		name:        name,
		RemotePorts: map[string]*InPort{},
		connected:   false,
	}
	return outp
}

// Name returns the name of the OutPort
func (pt *OutPort) Name() string {
	return pt.Process().Name() + "." + pt.name
}

// Process returns the process connected to the port
func (pt *OutPort) Process() WorkflowProcess {
	if pt.process == nil {
		Failf("Out-port %s has no connected process", pt.name)
	}
	return pt.process
}

// SetProcess sets the process of the port to p
func (pt *OutPort) SetProcess(p WorkflowProcess) {
	pt.process = p
}

// AddRemotePort adds a remote InPort to the OutPort
func (pt *OutPort) AddRemotePort(rpt *InPort) {
	if _, ok := pt.RemotePorts[rpt.Name()]; ok {
		Failf("[Process:%s]: A remote port with name %s already exists, for param-port %s connected to process %s\n", pt.Process().Name(), rpt.Name(), pt.Name(), rpt.Process().Name())
	}
	pt.RemotePorts[rpt.Name()] = rpt
}

// removeRemotePort removes the (in-)port with name rptName, from the OutPort
func (pt *OutPort) removeRemotePort(rptName string) {
	if _, ok := pt.RemotePorts[rptName]; !ok {
		Failf("[Process:%s]: No remote port with name %s exists, for param-port %s connected to process %s\n", pt.Process().Name(), rptName, pt.Name(), pt.Process().Name())
	}
	delete(pt.RemotePorts, rptName)
}

// Connect connects an InPort to the OutPort
func (pt *OutPort) Connect(rpt *InPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetConnectedStatus(true)
	rpt.SetConnectedStatus(true)
}

// Disconnect disconnects the (in-)port with name rptName, from the OutPort
func (pt *OutPort) Disconnect(rptName string) {
	pt.removeRemotePort(rptName)
	if len(pt.RemotePorts) == 0 {
		pt.SetConnectedStatus(false)
	}
}

// SetConnectedStatus sets the connected status of the OutPort
func (pt *OutPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

// Connected tells whether the port is connected or not
func (pt *OutPort) Connected() bool {
	return pt.connected
}

// Send sends an FileIP to all the in-ports connected to the OutPort
func (pt *OutPort) Send(ip *FileIP) {
	for _, rpt := range pt.RemotePorts {
		Debug.Printf("Sending on out-port %s connected to in-port %s", pt.Name(), rpt.Name())
		rpt.Send(ip)
	}
}

// Close closes the connection between this port and all the ports it is
// connected to. If this port is the last connected port to an in-port, that
// in-ports channel will also be closed.
func (pt *OutPort) Close() {
	for _, rpt := range pt.RemotePorts {
		Debug.Printf("Closing out-port %s connected to in-port %s", pt.Name(), rpt.Name())
		rpt.CloseConnection(pt.Name())
		pt.removeRemotePort(rpt.Name())
	}
}

// ------------------------------------------------------------------------
// ParamInPort
// ------------------------------------------------------------------------

// ParamInPort is an in-port for parameter values of string type
type ParamInPort struct {
	Chan        chan string
	name        string
	process     WorkflowProcess
	RemotePorts map[string]*ParamOutPort
	connected   bool
	closeLock   sync.Mutex
}

// NewParamInPort returns a new ParamInPort
func NewParamInPort(name string) *ParamInPort {
	return &ParamInPort{
		name:        name,
		Chan:        make(chan string, BUFSIZE),
		RemotePorts: map[string]*ParamOutPort{},
	}
}

// Name returns the name of the ParamInPort
func (pip *ParamInPort) Name() string {
	return pip.Process().Name() + "." + pip.name
}

// Process returns the process that is connected to the port
func (pip *ParamInPort) Process() WorkflowProcess {
	if pip.process == nil {
		Failf("Parameter in-port %s has no connected process", pip.name)
	}
	return pip.process
}

// SetProcess sets the process of the port to p
func (pip *ParamInPort) SetProcess(p WorkflowProcess) {
	pip.process = p
}

// AddRemotePort adds a remote ParamOutPort to the ParamInPort
func (pip *ParamInPort) AddRemotePort(pop *ParamOutPort) {
	if pip.RemotePorts[pop.Name()] != nil {
		Failf("[Process:%s]: A remote param port with name %s already exists, for in-param-port %s connected to process %s\n", pip.Process().Name(), pop.Name(), pip.Name(), pop.Process().Name())
	}
	pip.RemotePorts[pop.Name()] = pop
}

// Connect connects one parameter port with another one
func (pip *ParamInPort) Connect(pop *ParamOutPort) {
	pip.AddRemotePort(pop)
	pop.AddRemotePort(pip)

	pip.SetConnectedStatus(true)
	pop.SetConnectedStatus(true)
}

// ConnectStr connects a parameter port with a new go-routine feeding the
// strings in strings, on the fly, to the parameter port
func (pip *ParamInPort) ConnectStr(strings ...string) {
	pop := NewParamOutPort("string_feeder")
	pop.process = pip.Process()
	pip.Connect(pop)
	go func() {
		defer pop.Close()
		for _, str := range strings {
			pop.Send(str)
		}
	}()
}

// SetConnectedStatus sets the connected status of the ParamInPort
func (pip *ParamInPort) SetConnectedStatus(connected bool) {
	pip.connected = connected
}

// Connected tells whether the port is connected or not
func (pip *ParamInPort) Connected() bool {
	return pip.connected
}

// Send sends IPs to the in-port, and is supposed to be called from the remote
// (out-) port, to send to this in-port
func (pip *ParamInPort) Send(param string) {
	pip.Chan <- param
}

// Recv receiveds a param value over the ports connection
func (pip *ParamInPort) Recv() string {
	return <-pip.Chan
}

// CloseConnection closes the connection to the remote out-port with name
// popName, on the ParamInPort
func (pip *ParamInPort) CloseConnection(popName string) {
	pip.closeLock.Lock()
	delete(pip.RemotePorts, popName)
	if len(pip.RemotePorts) == 0 {
		close(pip.Chan)
	}
	pip.closeLock.Unlock()
}

// ------------------------------------------------------------------------
// ParamOutPort
// ------------------------------------------------------------------------

// ParamOutPort is an out-port for parameter values of string type
type ParamOutPort struct {
	name        string
	process     WorkflowProcess
	RemotePorts map[string]*ParamInPort
	connected   bool
}

// NewParamOutPort returns a new ParamOutPort
func NewParamOutPort(name string) *ParamOutPort {
	return &ParamOutPort{
		name:        name,
		RemotePorts: map[string]*ParamInPort{},
	}
}

// Name returns the name of the ParamOutPort
func (pop *ParamOutPort) Name() string {
	return pop.Process().Name() + "." + pop.name
}

// Process returns the process that is connected to the port
func (pop *ParamOutPort) Process() WorkflowProcess {
	if pop.process == nil {
		Failf("Parameter out-port %s has no connected process", pop.name)
	}
	return pop.process
}

// SetProcess sets the process of the port to p
func (pop *ParamOutPort) SetProcess(p WorkflowProcess) {
	pop.process = p
}

// AddRemotePort adds a remote ParamInPort to the ParamOutPort
func (pop *ParamOutPort) AddRemotePort(pip *ParamInPort) {
	if pop.RemotePorts[pip.Name()] != nil {
		Failf("[Process:%s]: A remote param port with name %s already exists, for in-param-port %s connected to process %s\n", pop.Process().Name(), pip.Name(), pop.Name(), pip.Process().Name())
	}
	pop.RemotePorts[pip.Name()] = pip
}

// Connect connects an ParamInPort to the ParamOutPort
func (pop *ParamOutPort) Connect(pip *ParamInPort) {
	pop.AddRemotePort(pip)
	pip.AddRemotePort(pop)

	pop.SetConnectedStatus(true)
	pip.SetConnectedStatus(true)
}

// Disconnect disonnects the (in-)port with name rptName, from the ParamOutPort
func (pop *ParamOutPort) Disconnect(pipName string) {
	pop.removeRemotePort(pipName)
	if len(pop.RemotePorts) == 0 {
		pop.SetConnectedStatus(false)
	}
}

// removeRemotePort removes the (in-)port with name rptName, from the ParamOutPort
func (pop *ParamOutPort) removeRemotePort(pipName string) {
	delete(pop.RemotePorts, pipName)
}

// SetConnectedStatus sets the connected status of the ParamOutPort
func (pop *ParamOutPort) SetConnectedStatus(connected bool) {
	pop.connected = connected
}

// Connected tells whether the port is connected or not
func (pop *ParamOutPort) Connected() bool {
	return pop.connected
}

// Send sends an FileIP to all the in-ports connected to the ParamOutPort
func (pop *ParamOutPort) Send(param string) {
	for _, pip := range pop.RemotePorts {
		Debug.Printf("Sending on out-param-port %s connected to in-param-port %s", pop.Name(), pip.Name())
		pip.Send(param)
	}
}

// Close closes the connection between this port and all the ports it is
// connected to. If this port is the last connected port to an in-port, that
// in-ports channel will also be closed.
func (pop *ParamOutPort) Close() {
	for _, pip := range pop.RemotePorts {
		Debug.Printf("Closing out-param-port %s connected to in-param-port %s", pop.Name(), pip.Name())
		pip.CloseConnection(pop.Name())
		pop.removeRemotePort(pip.Name())
	}
}
