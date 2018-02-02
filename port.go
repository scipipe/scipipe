package scipipe

func ConnectTo(outPort *OutPort, inPort *InPort) {
	outPort.Connect(inPort)
}

func ConnectFrom(inPort *InPort, outPort *OutPort) {
	outPort.Connect(inPort)
}

// Port is a struct that contains channels, together with some other meta data
// for keeping track of connection information between processes.
type InPort struct {
	name         string
	MergedInChan chan *IP
	inChans      []chan *IP
	connected    bool
}

func NewInPort(name string) *InPort {
	inp := &InPort{
		name:         name,
		MergedInChan: make(chan *IP, BUFSIZE), // This one will contain merged inputs from inChans
		inChans:      []chan *IP{},
		connected:    false,
	}
	return inp
}

func (pt *InPort) Name() string {
	return pt.name
}

func (localPort *InPort) Connect(remotePort *OutPort) {
	inBoundChan := make(chan *IP, BUFSIZE)
	localPort.AddInChan(inBoundChan)
	remotePort.AddOutChan(inBoundChan)

	localPort.SetConnectedStatus(true)
	remotePort.SetConnectedStatus(true)
}

func (pt *InPort) AddInChan(inChan chan *IP) {
	pt.inChans = append(pt.inChans, inChan)
}

func (pt *InPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

func (pt *InPort) IsConnected() bool {
	return pt.connected
}

// RunMerge merges (multiple) inputs on pt.inChans into pt.MergedInChan. This has to
// start running when the owning process runs, in order to merge in-ports
func (pt *InPort) RunMergeInputs() {
	defer close(pt.MergedInChan)
	for len(pt.inChans) > 0 {
		for i, ich := range pt.inChans {
			ip, ok := <-ich
			if !ok {
				// Delete in-channel at position i
				pt.inChans = append(pt.inChans[:i], pt.inChans[i+1:]...)
				break
			}
			pt.MergedInChan <- ip
		}
	}
}

func (pt *InPort) Recv() *IP {
	return <-pt.MergedInChan
}

// OutPort represents an output connection point on Processes
type OutPort struct {
	name      string
	outChans  []chan *IP
	connected bool
}

func NewOutPort(name string) *OutPort {
	outp := &OutPort{
		name:      name,
		outChans:  []chan *IP{},
		connected: false,
	}
	return outp
}

func (localPort *OutPort) Connect(remotePort *InPort) {
	outBoundChan := make(chan *IP, BUFSIZE)
	localPort.AddOutChan(outBoundChan)
	remotePort.AddInChan(outBoundChan)

	localPort.SetConnectedStatus(true)
	remotePort.SetConnectedStatus(true)
}

func (pt *OutPort) AddOutChan(outChan chan *IP) {
	pt.outChans = append(pt.outChans, outChan)
}

func (pt *OutPort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

func (pt *OutPort) IsConnected() bool {
	return pt.connected
}

func (pt *OutPort) Send(ip *IP) {
	for i, outChan := range pt.outChans {
		Debug.Printf("Sending on outchan %d in port\n", i)
		outChan <- ip
	}
}

func (pt *OutPort) Close() {
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
