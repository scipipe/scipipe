package scipipe

// Sink is a simple component that just receives IPs on its In-port without
// doing anything with them. It is used to drive pipelines of processes
type Sink struct {
	EmptyWorkflowProcess
	name        string
	inPort      *InPort
	paramInPort *ParamInPort
}

// NewSink returns a new Sink component
func NewSink(name string) (s *Sink) {
	inp := NewInPort("sink_in")
	pip := NewParamInPort("sink_param_in")
	snk := &Sink{
		name:        name,
		inPort:      inp,
		paramInPort: pip,
	}
	inp.Process = snk
	pip.Process = snk
	return snk
}

// IsConnected checks whether the sinks in-port is connected
func (p *Sink) IsConnected() bool {
	return p.inPort.IsConnected() || p.paramInPort.IsConnected()
}

// Connect connects an out-port to the sinks in-port
func (p *Sink) Connect(outPort *OutPort) {
	p.inPort.Connect(outPort)
}

// ConnectParam connects a param-out-port to the sinks param-in-port
func (p *Sink) ConnectParam(paramOutPort *ParamOutPort) {
	p.paramInPort.Connect(paramOutPort)
}

// Name returns the name of the sink
func (p *Sink) Name() string {
	return p.name
}

// Run runs the Sink process
func (p *Sink) Run() {
	merged := make(chan int)
	if p.inPort.IsConnected() {
		go func() {
			for ip := range p.inPort.Chan {
				Debug.Printf("Got file in sink: %s\n", ip.Path())
			}
			merged <- 1
		}()
	}
	if p.paramInPort.IsConnected() {
		go func() {
			for param := range p.paramInPort.Chan {
				Debug.Printf("Got param in sink: %s\n", param)
			}
			merged <- 1
		}()
	}
	if p.inPort.IsConnected() {
		<-merged
	}
	if p.paramInPort.IsConnected() {
		<-merged
	}
	close(merged)
	Debug.Printf("Caught up everything in sink")
}
