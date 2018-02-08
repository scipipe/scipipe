package scipipe

// Sink is a simple component that just receives IPs on its In-port without
// doing anything with them. It is used to drive pipelines of processes
type Sink struct {
	BaseProcess
}

// NewSink returns a new Sink component
func NewSink(wf *Workflow, name string) *Sink {
	p := &Sink{
		BaseProcess: NewBaseProcess(wf, name),
	}
	p.InitInPort(p, "sink_in")
	p.InitParamInPort(p, "param_sink_in")
	return p
}

func (p *Sink) in() *InPort           { return p.InPort("sink_in") }
func (p *Sink) paramIn() *ParamInPort { return p.ParamInPort("param_sink_in") }

// Connect connects an out-port to the sinks in-port
func (p *Sink) Connect(outPort *OutPort) {
	p.in().Connect(outPort)
}

// ConnectParam connects a param-out-port to the sinks param-in-port
func (p *Sink) ConnectParam(paramOutPort *ParamOutPort) {
	p.paramIn().Connect(paramOutPort)
}

// Run runs the Sink process
func (p *Sink) Run() {
	merged := make(chan int)
	if p.in().Connected() {
		go func() {
			for ip := range p.in().Chan {
				Debug.Printf("Got file in sink: %s\n", ip.Path())
			}
			merged <- 1
		}()
	}
	if p.paramIn().Connected() {
		go func() {
			for param := range p.paramIn().Chan {
				Debug.Printf("Got param in sink: %s\n", param)
			}
			merged <- 1
		}()
	}
	if p.in().Connected() {
		<-merged
	}
	if p.paramIn().Connected() {
		<-merged
	}
	close(merged)
	Debug.Printf("Caught up everything in sink")
}
