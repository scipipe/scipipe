package components

import (
	"github.com/scipipe/scipipe"
)

// StreamToSubStream takes a normal stream of IP's representing
// individual files, and returns one IP where the incoming IPs
// are sent on its substream.
type StreamToSubStream struct {
	scipipe.EmptyWorkflowProcess
	name         string
	In           *scipipe.InPort
	OutSubStream *scipipe.OutPort
}

// NewStreamToSubStream instantiates a new StreamToSubStream process
func NewStreamToSubStream(wf *scipipe.Workflow, name string) *StreamToSubStream {
	stss := &StreamToSubStream{
		name:         name,
		In:           scipipe.NewInPort("in"),
		OutSubStream: scipipe.NewOutPort("out_substream"),
	}
	wf.AddProc(stss)
	return stss
}

// Name returns the name of the StreamToSubStream process
func (p *StreamToSubStream) Name() string {
	return p.name
}

// Connected tells whether all the ports of the process are connected
func (p *StreamToSubStream) Connected() bool {
	return p.In.Connected() && p.OutSubStream.Connected()
}

// InPorts returns all the in-ports for the process
func (p *StreamToSubStream) InPorts() map[string]*scipipe.InPort {
	return map[string]*scipipe.InPort{
		p.In.Name(): p.In,
	}
}

// OutPorts returns all the out-ports for the process
func (p *StreamToSubStream) OutPorts() map[string]*scipipe.OutPort {
	return map[string]*scipipe.OutPort{
		p.OutSubStream.Name(): p.OutSubStream,
	}
}

// Run runs the StreamToSubStream
func (p *StreamToSubStream) Run() {
	defer p.OutSubStream.Close()

	scipipe.Debug.Println("Creating new information packet for the substream...")
	subStreamIP := scipipe.NewIP("")
	scipipe.Debug.Printf("Setting in-port of process %s to IP substream field\n", p.Name())
	subStreamIP.SubStream = p.In

	scipipe.Debug.Printf("Sending sub-stream IP in process %s...\n", p.Name())
	p.OutSubStream.Send(subStreamIP)
	scipipe.Debug.Printf("Done sending sub-stream IP in process %s.\n", p.Name())
}
