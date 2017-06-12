package scipipe

import (
	"os"
)

type Port interface {
	Connect(Port)
	IsConnected() bool
	SetConnectedStatus(bool)
}

func Connect(port1 *FilePort, port2 *FilePort) {
	port1.Connect(port2)
}

// FilePort
type FilePort struct {
	Port
	Chan        chan *InformationPacket // TODO: Deprecated
	remoteChans []chan *InformationPacket
	connected   bool
}

func NewFilePort() *FilePort {
	return &FilePort{
		Chan:        make(chan *InformationPacket, BUFSIZE),
		remoteChans: []chan *InformationPacket{},
		connected:   false,
	}

}

func (localPort *FilePort) Connect(remotePort *FilePort) {
	// Needed to make this work as an in-port
	localPort.Chan = remotePort.Chan

	localPort.AddRemoteChan(remotePort.Chan)
	remotePort.AddRemoteChan(localPort.Chan)

	localPort.SetConnectedStatus(true)
	remotePort.SetConnectedStatus(true)
}

func (pt *FilePort) AddRemoteChan(remoteChan chan *InformationPacket) {
	pt.remoteChans = append(pt.remoteChans, remoteChan)
}

func (pt *FilePort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

func (pt *FilePort) IsConnected() bool {
	return pt.connected
}

func (pt *FilePort) Send(ip *InformationPacket) {
	for _, remoteChan := range pt.remoteChans {
		remoteChan <- ip
	}
}

func (pt *FilePort) Recv() *InformationPacket {
	return <-pt.Chan // TODO: Support more than one in-port too
}

func (pt *FilePort) Close() {
	for _, remoteChan := range pt.remoteChans {
		close(remoteChan)
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
		Error.Println("Both paramports already have initialized channels, so can't choose which to use!")
		os.Exit(1)
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
