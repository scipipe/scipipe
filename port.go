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
	ready       bool
	closeLock   sync.Mutex
}

// NewInPort returns a new InPort struct
func NewInPort(name string) *InPort {
	inp := &InPort{
		name:        name,
		RemotePorts: map[string]*OutPort{},
		Chan:        make(chan *FileIP, BUFSIZE), // This one will contain merged inputs from inChans
		ready:       false,
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

// From connects an OutPort to the InPort
func (pt *InPort) From(rpt *OutPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetReady(true)
	rpt.SetReady(true)
}

// Disconnect disconnects the (out-)port with name rptName, from the InPort
func (pt *InPort) Disconnect(rptName string) {
	pt.removeRemotePort(rptName)
	if len(pt.RemotePorts) == 0 {
		pt.SetReady(false)
	}
}

// removeRemotePort removes the (out-)port with name rptName, from the InPort
func (pt *InPort) removeRemotePort(rptName string) {
	delete(pt.RemotePorts, rptName)
}

// SetReady sets the ready status of the InPort
func (pt *InPort) SetReady(ready bool) {
	pt.ready = ready
}

// Ready tells whether the port is ready or not
func (pt *InPort) Ready() bool {
	return pt.ready
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
	ready       bool
}

// NewOutPort returns a new OutPort struct
func NewOutPort(name string) *OutPort {
	outp := &OutPort{
		name:        name,
		RemotePorts: map[string]*InPort{},
		ready:       false,
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

// To connects an InPort to the OutPort
func (pt *OutPort) To(rpt *InPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetReady(true)
	rpt.SetReady(true)
}

// Disconnect disconnects the (in-)port with name rptName, from the OutPort
func (pt *OutPort) Disconnect(rptName string) {
	pt.removeRemotePort(rptName)
	if len(pt.RemotePorts) == 0 {
		pt.SetReady(false)
	}
}

// SetReady sets the ready status of the OutPort
func (pt *OutPort) SetReady(ready bool) {
	pt.ready = ready
}

// Ready tells whether the port is ready or not
func (pt *OutPort) Ready() bool {
	return pt.ready
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
// InParamPort
// ------------------------------------------------------------------------

// InParamPort is an in-port for parameter values of string type
type InParamPort struct {
	Chan        chan string
	name        string
	process     WorkflowProcess
	RemotePorts map[string]*OutParamPort
	ready       bool
	closeLock   sync.Mutex
}

// NewInParamPort returns a new InParamPort
func NewInParamPort(name string) *InParamPort {
	return &InParamPort{
		name:        name,
		Chan:        make(chan string, BUFSIZE),
		RemotePorts: map[string]*OutParamPort{},
	}
}

// Name returns the name of the InParamPort
func (pip *InParamPort) Name() string {
	return pip.Process().Name() + "." + pip.name
}

// Process returns the process that is connected to the port
func (pip *InParamPort) Process() WorkflowProcess {
	if pip.process == nil {
		Failf("Parameter in-port %s has no connected process", pip.name)
	}
	return pip.process
}

// SetProcess sets the process of the port to p
func (pip *InParamPort) SetProcess(p WorkflowProcess) {
	pip.process = p
}

// AddRemotePort adds a remote OutParamPort to the InParamPort
func (pip *InParamPort) AddRemotePort(pop *OutParamPort) {
	if pip.RemotePorts[pop.Name()] != nil {
		Failf("[Process:%s]: A remote param port with name %s already exists, for in-param-port %s connected to process %s\n", pip.Process().Name(), pop.Name(), pip.Name(), pop.Process().Name())
	}
	pip.RemotePorts[pop.Name()] = pop
}

// From connects one parameter port with another one
func (pip *InParamPort) From(pop *OutParamPort) {
	pip.AddRemotePort(pop)
	pop.AddRemotePort(pip)

	pip.SetReady(true)
	pop.SetReady(true)
}

// FromStr connects a parameter port with a new go-routine feeding the
// strings in strings, on the fly, to the parameter port
func (pip *InParamPort) FromStr(strings ...string) {
	pop := NewOutParamPort("string_feeder")
	pop.process = pip.Process()
	pip.From(pop)
	go func() {
		defer pop.Close()
		for _, str := range strings {
			pop.Send(str)
		}
	}()
}

// SetReady sets the ready status of the InParamPort
func (pip *InParamPort) SetReady(ready bool) {
	pip.ready = ready
}

// Ready tells whether the port is ready or not
func (pip *InParamPort) Ready() bool {
	return pip.ready
}

// Send sends IPs to the in-port, and is supposed to be called from the remote
// (out-) port, to send to this in-port
func (pip *InParamPort) Send(param string) {
	pip.Chan <- param
}

// Recv receiveds a param value over the ports connection
func (pip *InParamPort) Recv() string {
	return <-pip.Chan
}

// CloseConnection closes the connection to the remote out-port with name
// popName, on the InParamPort
func (pip *InParamPort) CloseConnection(popName string) {
	pip.closeLock.Lock()
	delete(pip.RemotePorts, popName)
	if len(pip.RemotePorts) == 0 {
		close(pip.Chan)
	}
	pip.closeLock.Unlock()
}

// ------------------------------------------------------------------------
// OutParamPort
// ------------------------------------------------------------------------

// OutParamPort is an out-port for parameter values of string type
type OutParamPort struct {
	name        string
	process     WorkflowProcess
	RemotePorts map[string]*InParamPort
	ready       bool
}

// NewOutParamPort returns a new OutParamPort
func NewOutParamPort(name string) *OutParamPort {
	return &OutParamPort{
		name:        name,
		RemotePorts: map[string]*InParamPort{},
	}
}

// Name returns the name of the OutParamPort
func (pop *OutParamPort) Name() string {
	return pop.Process().Name() + "." + pop.name
}

// Process returns the process that is connected to the port
func (pop *OutParamPort) Process() WorkflowProcess {
	if pop.process == nil {
		Failf("Parameter out-port %s has no connected process", pop.name)
	}
	return pop.process
}

// SetProcess sets the process of the port to p
func (pop *OutParamPort) SetProcess(p WorkflowProcess) {
	pop.process = p
}

// AddRemotePort adds a remote InParamPort to the OutParamPort
func (pop *OutParamPort) AddRemotePort(pip *InParamPort) {
	if pop.RemotePorts[pip.Name()] != nil {
		Failf("[Process:%s]: A remote param port with name %s already exists, for in-param-port %s connected to process %s\n", pop.Process().Name(), pip.Name(), pop.Name(), pip.Process().Name())
	}
	pop.RemotePorts[pip.Name()] = pip
}

// To connects an InParamPort to the OutParamPort
func (pop *OutParamPort) To(pip *InParamPort) {
	pop.AddRemotePort(pip)
	pip.AddRemotePort(pop)

	pop.SetReady(true)
	pip.SetReady(true)
}

// Disconnect disonnects the (in-)port with name rptName, from the OutParamPort
func (pop *OutParamPort) Disconnect(pipName string) {
	pop.removeRemotePort(pipName)
	if len(pop.RemotePorts) == 0 {
		pop.SetReady(false)
	}
}

// removeRemotePort removes the (in-)port with name rptName, from the OutParamPort
func (pop *OutParamPort) removeRemotePort(pipName string) {
	delete(pop.RemotePorts, pipName)
}

// SetReady sets the ready status of the OutParamPort
func (pop *OutParamPort) SetReady(ready bool) {
	pop.ready = ready
}

// Ready tells whether the port is ready or not
func (pop *OutParamPort) Ready() bool {
	return pop.ready
}

// Send sends an FileIP to all the in-ports connected to the OutParamPort
func (pop *OutParamPort) Send(param string) {
	for _, pip := range pop.RemotePorts {
		Debug.Printf("Sending on out-param-port %s connected to in-param-port %s", pop.Name(), pip.Name())
		pip.Send(param)
	}
}

// Close closes the connection between this port and all the ports it is
// connected to. If this port is the last connected port to an in-port, that
// in-ports channel will also be closed.
func (pop *OutParamPort) Close() {
	for _, pip := range pop.RemotePorts {
		Debug.Printf("Closing out-param-port %s connected to in-param-port %s", pop.Name(), pip.Name())
		pip.CloseConnection(pop.Name())
		pop.removeRemotePort(pip.Name())
	}
}
