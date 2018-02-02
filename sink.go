package scipipe

// Sink is a simple component that just receives IPs on its In-port without
// doing anything with them. It is used to drive pipelines of processes
type Sink struct {
	name   string
	inPort *InPort
}

// NewSink returns a new Sink component
func NewSink(name string) (s *Sink) {
	return &Sink{
		name:   name,
		inPort: NewInPort("in"),
	}
}

// IsConnected checks whether the sinks in-port is connected
func (p *Sink) IsConnected() bool {
	return p.inPort.IsConnected()
}

// Connect connects and out-port to the sinks in-port
func (p *Sink) Connect(outPort *OutPort) {
	p.inPort.Connect(outPort)
}

// Name returns the name of the sink
func (p *Sink) Name() string {
	return p.name
}

// Run runs the Sink process
func (p *Sink) Run() {
	for ip := range p.inPort.Chan {
		Debug.Printf("Got file in sink: %s\n", ip.GetPath())
	}
}
