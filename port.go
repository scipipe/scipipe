package scipipe

import (
	"sync"
)

// ConnectTo connects an OutPort to an InPort
func ConnectTo(outPort *OutPort, inPort *InPort) {
	outPort.Connect(inPort)
}

// ConnectFrom connects from an InPort to an OutPort
func ConnectFrom(inPort *InPort, outPort *OutPort) {
	outPort.Connect(inPort)
}

// InPort represents a pluggable connection to multiple out-ports from other
// processes, from its own process, and with which it is communicating via
// channels under the hood
type InPort struct {
	Chan        chan *IP
	name        string
	Process     WorkflowProcess
	RemotePorts map[string]*OutPort
	connected   bool
	closeLock   sync.Mutex
}

// NewInPort returns a new InPort struct
func NewInPort(name string) *InPort {
	inp := &InPort{
		name:        name,
		RemotePorts: map[string]*OutPort{},
		Chan:        make(chan *IP, BUFSIZE), // This one will contain merged inputs from inChans
		connected:   false,
	}
	return inp
}

// Name returns the name of the InPort
func (pt *InPort) Name() string {
	if pt.Process != nil {
		return pt.Process.Name() + "." + pt.name
	}
	return pt.name
}

// AddRemotePort adds a remote OutPort to the InPort
func (pt *InPort) AddRemotePort(rpt *OutPort) {
	if pt.RemotePorts[rpt.Name()] != nil {
		Error.Fatalf("A remote port with name %s already exists, for in-port %s\n", rpt.Name(), pt.Name())
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

// SetConnectedStatus sets the connected status of the InPort
func (pt *InPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

// IsConnected tells whether the port is connected or not
func (pt *InPort) IsConnected() bool {
	return pt.connected
}

// Send sends IPs to the in-port, and is supposed to be called from the remote
// (out-) port, to send to this in-port
func (pt *InPort) Send(ip *IP) {
	pt.Chan <- ip
}

// Recv receives IPs from the port
func (pt *InPort) Recv() *IP {
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

// OutPort represents a pluggable connection to multiple in-ports from other
// processes, from its own process, and with which it is communicating via
// channels under the hood
type OutPort struct {
	name        string
	Process     WorkflowProcess
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
	if pt.Process != nil {
		return pt.Process.Name() + "." + pt.name
	}
	return pt.name
}

// AddRemotePort adds a remote InPort to the OutPort
func (pt *OutPort) AddRemotePort(rpt *InPort) {
	if pt.RemotePorts[rpt.Name()] != nil {
		Error.Fatalf("A remote port with name %s already exists, for out-port %s\n", rpt.Name(), pt.Name())
	}
	pt.RemotePorts[rpt.Name()] = rpt
}

// RemoveRemotePort removes the (in-)port with name rptName, from the OutPort
func (pt *OutPort) RemoveRemotePort(rptName string) {
	delete(pt.RemotePorts, rptName)
}

// Connect connects an InPort to the OutPort
func (pt *OutPort) Connect(rpt *InPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetConnectedStatus(true)
	rpt.SetConnectedStatus(true)
}

// SetConnectedStatus sets the connected status of the OutPort
func (pt *OutPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

// IsConnected tells whether the port is connected or not
func (pt *OutPort) IsConnected() bool {
	return pt.connected
}

// Send sends an IP to all the in-ports connected to the OutPort
func (pt *OutPort) Send(ip *IP) {
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
		pt.RemoveRemotePort(rpt.Name())
	}
}

// ParamPort is a port for parameter values of string type
type ParamPort struct {
	Chan      chan string
	connected bool
}

// NewParamPort returns a new ParamPort
func NewParamPort() *ParamPort {
	return &ParamPort{}
}

// Connect connects one parameter port with another one
func (pp *ParamPort) Connect(otherParamPort *ParamPort) {
	if pp.Chan != nil && otherParamPort.Chan != nil {
		Error.Fatalln("Both paramports already have initialized channels, so can't choose which to use!")
	} else if pp.Chan != nil && otherParamPort.Chan == nil {
		Debug.Println("Local param port, but not the other one, initialized, so connecting local to other")
		otherParamPort.Chan = pp.Chan
	} else if otherParamPort.Chan != nil && pp.Chan == nil {
		Debug.Println("The other, but not the local param port initialized, so connecting other to local")
		pp.Chan = otherParamPort.Chan
	} else if pp.Chan == nil && otherParamPort.Chan == nil {
		Debug.Println("Neither local nor other param port initialized, so creating new channel and connecting both")
		ch := make(chan string, BUFSIZE)
		pp.Chan = ch
		otherParamPort.Chan = ch
	}
	pp.SetConnectedStatus(true)
	otherParamPort.SetConnectedStatus(true)
}

// ConnectStr connects a parameter port with a new go-routine feeding the
// strings in strings, on the fly, to the parameter port
func (pp *ParamPort) ConnectStr(strings ...string) {
	pp.Chan = make(chan string, BUFSIZE)
	pp.SetConnectedStatus(true)
	go func() {
		defer pp.Close()
		for _, str := range strings {
			pp.Chan <- str
		}
	}()
}

// SetConnectedStatus sets the connected status of the ParamPort
func (pp *ParamPort) SetConnectedStatus(connected bool) {
	pp.connected = connected
}

// IsConnected tells whether the port is connected or not
func (pp *ParamPort) IsConnected() bool {
	return pp.connected
}

// Send sends the param value over the ports connection
func (pp *ParamPort) Send(param string) {
	pp.Chan <- param
}

// Recv receiveds a param value over the ports connection
func (pp *ParamPort) Recv() string {
	return <-pp.Chan
}

// Close closes the port (and its channel)
func (pp *ParamPort) Close() {
	close(pp.Chan)
}
