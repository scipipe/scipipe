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
	p.InitInParamPort(p, "param_sink_in")
	return p
}

func (p *Sink) in() *InPort           { return p.InPort("sink_in") }
func (p *Sink) paramIn() *InParamPort { return p.InParamPort("param_sink_in") }

// From connects an out-port to the sinks in-port
func (p *Sink) From(outPort *OutPort) {
	p.in().From(outPort)
}

// FromParam connects a param-out-port to the sinks param-in-port
func (p *Sink) FromParam(outParamPort *OutParamPort) {
	p.paramIn().From(outParamPort)
}

// Run runs the Sink process
func (p *Sink) Run() {
	merged := make(chan int)
	if p.in().Ready() {
		go func() {
			for ip := range p.in().Chan {
				Debug.Printf("Got file in sink: %s\n", ip.Path())
			}
			merged <- 1
		}()
	}
	if p.paramIn().Ready() {
		go func() {
			for param := range p.paramIn().Chan {
				Debug.Printf("Got param in sink: %s\n", param)
			}
			merged <- 1
		}()
	}
	if p.in().Ready() {
		<-merged
	}
	if p.paramIn().Ready() {
		<-merged
	}
	close(merged)
	Debug.Printf("Caught up everything in sink")
}
