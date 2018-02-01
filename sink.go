package scipipe

// Sink is a simple component that just receives IP on its In-port
// without doing anything with them
type Sink struct {
	Process
	name   string
	inPort *Port
}

// Instantiate a Sink component
func NewSink(name string) (s *Sink) {
	return &Sink{
		name:   name,
		inPort: NewPort(),
	}
}

func (p *Sink) IsConnected() bool {
	return p.inPort.IsConnected()
}

func (p *Sink) Connect(outPort *Port) {
	p.inPort.Connect(outPort)
}

func (p *Sink) Name() string {
	return p.name
}

// Execute the Sink component
func (p *Sink) Run() {
	go p.inPort.RunMergeInputs()
	for ip := range p.inPort.InChan {
		Debug.Printf("Got file in sink: %s\n", ip.GetPath())
	}
}
