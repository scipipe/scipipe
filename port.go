package scipipe

import (
	"sync"
)

func ConnectTo(outPort *OutPort, inPort *InPort) {
	outPort.Connect(inPort)
}

func ConnectFrom(inPort *InPort, outPort *OutPort) {
	outPort.Connect(inPort)
}

// Port is a struct that contains channels, together with some other meta data
// for keeping track of connection information between processes.
type InPort struct {
	Chan        chan *IP
	name        string
	Process     WorkflowProcess
	RemotePorts map[string]*OutPort
	connected   bool
	closeLock   sync.Mutex
}

func NewInPort(name string) *InPort {
	inp := &InPort{
		name:        name,
		RemotePorts: map[string]*OutPort{},
		Chan:        make(chan *IP, BUFSIZE), // This one will contain merged inputs from inChans
		connected:   false,
	}
	return inp
}

func (pt *InPort) Name() string {
	if pt.Process != nil {
		return pt.Process.Name() + "." + pt.name
	}
	return pt.name
}

func (pt *InPort) AddRemotePort(rpt *OutPort) {
	if pt.RemotePorts[rpt.Name()] != nil {
		Error.Fatalf("A remote port with name %s already exists, for in-port %s\n", rpt.Name(), pt.Name())
	}
	pt.RemotePorts[rpt.Name()] = rpt
}

func (pt *InPort) Connect(rpt *OutPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetConnectedStatus(true)
	rpt.SetConnectedStatus(true)
}

func (pt *InPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

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

func (pt *InPort) CloseConnection(rptName string) {
	pt.closeLock.Lock()
	delete(pt.RemotePorts, rptName)
	if len(pt.RemotePorts) == 0 {
		close(pt.Chan)
	}
	pt.closeLock.Unlock()
}

// OutPort represents an output connection point on Processes
type OutPort struct {
	name        string
	Process     WorkflowProcess
	RemotePorts map[string]*InPort
	connected   bool
}

func NewOutPort(name string) *OutPort {
	outp := &OutPort{
		name:        name,
		RemotePorts: map[string]*InPort{},
		connected:   false,
	}
	return outp
}

func (pt *OutPort) Name() string {
	if pt.Process != nil {
		return pt.Process.Name() + "." + pt.name
	}
	return pt.name
}

func (pt *OutPort) AddRemotePort(rpt *InPort) {
	pt.RemotePorts[rpt.Name()] = rpt
}

func (pt *OutPort) RemoveRemotePort(rptName string) {
	delete(pt.RemotePorts, rptName)
}

func (pt *OutPort) Connect(rpt *InPort) {
	pt.AddRemotePort(rpt)
	rpt.AddRemotePort(pt)

	pt.SetConnectedStatus(true)
	rpt.SetConnectedStatus(true)
}

func (pt *OutPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

func (pt *OutPort) IsConnected() bool {
	return pt.connected
}

func (pt *OutPort) Send(ip *IP) {
	for _, rpt := range pt.RemotePorts {
		Debug.Printf("Sending on out-port %s connected to in-port %s", pt.Name(), rpt.Name())
		rpt.Send(ip)
	}
}

func (pt *OutPort) Close() {
	for _, rpt := range pt.RemotePorts {
		Debug.Printf("Closing out-port %s connected to in-port %s", pt.Name(), rpt.Name())
		rpt.CloseConnection(pt.Name())
		pt.RemoveRemotePort(rpt.Name())
	}
}

// ParamPort
type ParamPort struct {
	Chan      chan string
	connected bool
}

func NewParamPort() *ParamPort {
	return &ParamPort{}
}

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

func (pp *ParamPort) SetConnectedStatus(connected bool) {
	pp.connected = connected
}

func (pp *ParamPort) IsConnected() bool {
	return pp.connected
}

func (pp *ParamPort) Send(param string) {
	pp.Chan <- param
}

func (pp *ParamPort) Recv() string {
	return <-pp.Chan
}

func (pp *ParamPort) Close() {
	close(pp.Chan)
}
