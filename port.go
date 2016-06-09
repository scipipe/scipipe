package scipipe

type Port interface {
	Connect(Port)
	IsConnected() bool
	SetConnectedStatus(bool)
}

func ConnectFromTo(outPort *OutPort, inPort *InPort) {
	inPort.Connect(outPort)
}

func ConnectToFrom(inPort *InPort, outPort *OutPort) {
	inPort.Connect(outPort)
}

// InPort
type InPort struct {
	Port
	Chan      chan *FileTarget
	connected bool
}

func NewInPort() *InPort {
	return &InPort{connected: false}
}

func (inp *InPort) Connect(outp *OutPort) {
	if inp.Chan != nil && outp.Chan != nil {
		Error.Println("Both in-port and out-port already have initialized channels, so can't choose which to use!")
	} else if inp.Chan != nil && outp.Chan == nil {
		Debug.Println("InPort, but not OutPort initialized, so connecting InPort to OutPort")
		outp.Chan = inp.Chan
	} else if outp.Chan != nil && inp.Chan == nil {
		Debug.Println("OutPort, but not InPort initialized, so connecting OutPort to InPort")
		inp.Chan = outp.Chan
	} else if inp.Chan == nil && outp.Chan == nil {
		Debug.Println("Neither InPort nor OutPort initialized, so creating new channel")
		ch := make(chan *FileTarget, BUFSIZE)
		inp.Chan = ch
		outp.Chan = ch
	}
	inp.SetConnectedStatus(true)
	outp.SetConnectedStatus(true)
}

func (inp *InPort) SetConnectedStatus(connected bool) {
	inp.connected = connected
}

func (inp *InPort) IsConnected() bool {
	return inp.connected
}

// OutPort
type OutPort struct {
	Port
	Chan      chan *FileTarget
	connected bool
}

func NewOutPort() *OutPort {
	return &OutPort{connected: false}
}

func (outp *OutPort) Connect(inp *InPort) {
	inp.Connect(outp)
}

func (outp *OutPort) IsConnected() bool {
	return outp.connected
}

func (outp *OutPort) SetConnectedStatus(connected bool) {
	outp.connected = connected
}

func (outp *OutPort) Close() {
	close(outp.Chan)
}

// ParamPort
type ParamPort struct {
	Chan      chan string
	connected bool
}

func NewParamPort() *ParamPort {
	return &ParamPort{}
}

func (paramp *ParamPort) Connect(otherParamPort *ParamPort) {
	if paramp.Chan != nil && otherParamPort.Chan != nil {
		Error.Println("Both paramports already have initialized channels, so can't choose which to use!")
	} else if paramp.Chan != nil && otherParamPort.Chan == nil {
		Debug.Println("Local param port, but not the other one, initialized, so connecting local to other")
		otherParamPort.Chan = paramp.Chan
	} else if otherParamPort.Chan != nil && paramp.Chan == nil {
		Debug.Println("The other, but not the local param port initialized, so connecting other to local")
		paramp.Chan = otherParamPort.Chan
	} else if paramp.Chan == nil && otherParamPort.Chan == nil {
		Debug.Println("Neither local nor other param port initialized, so creating new channel and connecting both")
		ch := make(chan string, BUFSIZE)
		paramp.Chan = ch
		otherParamPort.Chan = ch
	}
	paramp.SetConnectedStatus(true)
	otherParamPort.SetConnectedStatus(true)
}

func (paramp *ParamPort) SetConnectedStatus(connected bool) {
	paramp.connected = connected
}

func (paramp *ParamPort) IsConnected() bool {
	return paramp.connected
}

func (paramp *ParamPort) Close() {
	close(paramp.Chan)
}
