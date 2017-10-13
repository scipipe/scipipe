package scipipe

// Sink is a simple component that just receives InformationPacket on its In-port
// without doing anything with them
type Sink struct {
	Process
	name   string
	inPort *FilePort
}

// Instantiate a Sink component
func NewSink(name string) (s *Sink) {
	return &Sink{
		name:   name,
		inPort: NewFilePort(),
	}
}

func (p *Sink) IsConnected() bool {
	return p.inPort.IsConnected()
}

func (p *Sink) Connect(outPort *FilePort) {
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
