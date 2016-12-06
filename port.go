package scipipe

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
	Chan      chan *FileTarget
	connected bool
}

func NewFilePort() *FilePort {
	return &FilePort{connected: false}
}

func (pt1 *FilePort) Connect(pt2 *FilePort) {
	if pt1.Chan != nil && pt2.Chan != nil {
		Error.Println("Both ports already have initialized channels, so can't choose which to use!")
	} else if pt1.Chan != nil && pt2.Chan == nil {
		Debug.Println("port2 not initialized, so connecting port1 to port2")
		pt2.Chan = pt1.Chan
	} else if pt2.Chan != nil && pt1.Chan == nil {
		Debug.Println("port1 not initialized, so connecting port2 to port1")
		pt1.Chan = pt2.Chan
	} else if pt1.Chan == nil && pt2.Chan == nil {
		Debug.Println("Neither port1 nor port2 initialized, so creating new channel")
		ch := make(chan *FileTarget, BUFSIZE)
		pt1.Chan = ch
		pt2.Chan = ch
	}
	pt1.SetConnectedStatus(true)
	pt2.SetConnectedStatus(true)
}

func (pt *FilePort) SetConnectedStatus(connected bool) {
	pt.connected = connected
}

func (pt *FilePort) IsConnected() bool {
	return pt.connected
}

func (pt *FilePort) Close() {
	close(pt.Chan)
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
