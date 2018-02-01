package scipipe

import (
	"os"
)

func Connect(port1 *Port, port2 *Port) {
	port1.Connect(port2)
}

// Port
type Port struct {
	InChan    chan *IP
	inChans   []chan *IP
	outChans  []chan *IP
	connected bool
}

func NewPort() *Port {
	fp := &Port{
		InChan:    make(chan *IP, BUFSIZE), // This one will contain merged inputs from inChans
		inChans:   []chan *IP{},
		outChans:  []chan *IP{},
		connected: false,
	}
	return fp
}

func (localPort *Port) Connect(remotePort *Port) {
	// If localPort is an in-port
	inBoundChan := make(chan *IP, BUFSIZE)
	localPort.AddInChan(inBoundChan)
	remotePort.AddOutChan(inBoundChan)

	// If localPort is an out-port
	outBoundChan := make(chan *IP, BUFSIZE)
	localPort.AddOutChan(outBoundChan)
	remotePort.AddInChan(outBoundChan)

	localPort.SetConnectedStatus(true)
	remotePort.SetConnectedStatus(true)
}

// RunMerge merges (multiple) inputs on pt.inChans into pt.InChan. This has to
// start running when the owning process runs, in order to merge in-ports
func (pt *Port) RunMergeInputs() {
	defer close(pt.InChan)
	for len(pt.inChans) > 0 {
		for i, ich := range pt.inChans {
			ip, ok := <-ich
			if !ok {
				// Delete in-channel at position i
				pt.inChans = append(pt.inChans[:i], pt.inChans[i+1:]...)
				break
			}
			pt.InChan <- ip
		}
	}
}

func (pt *Port) AddOutChan(outChan chan *IP) {
	pt.outChans = append(pt.outChans, outChan)
}

func (pt *Port) AddInChan(inChan chan *IP) {
	pt.inChans = append(pt.inChans, inChan)
}

func (pt *Port) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

func (pt *Port) IsConnected() bool {
	return pt.connected
}

func (pt *Port) Send(ip *IP) {
	for i, outChan := range pt.outChans {
		Debug.Printf("Sending on outchan %d in port\n", i)
		outChan <- ip
	}
}

func (pt *Port) Recv() *IP {
	return <-pt.InChan
}

func (pt *Port) Close() {
	for i, outChan := range pt.outChans {
		Debug.Printf("Closing outchan %d in port\n", i)
		close(outChan)
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
