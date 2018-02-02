package scipipe

// Sink is a simple component that just receives IP on its In-port
// without doing anything with them
type Sink struct {
	name   string
	inPort *InPort
}

// Instantiate a Sink component
func NewSink(name string) (s *Sink) {
	return &Sink{
		name:   name,
		inPort: NewInPort("in"),
	}
}

func (p *Sink) IsConnected() bool {
	return p.inPort.IsConnected()
}

func (p *Sink) Connect(outPort *OutPort) {
	p.inPort.Connect(outPort)
}

func (p *Sink) Name() string {
	return p.name
}

// Execute the Sink component
func (p *Sink) Run() {
	for ip := range p.inPort.Chan {
		Debug.Printf("Got file in sink: %s\n", ip.GetPath())
	}
}
