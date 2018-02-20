package components

import (
	"github.com/scipipe/scipipe"
)

// StreamToSubStream takes a normal stream of IP's representing
// individual files, and returns one IP where the incoming IPs
// are sent on its substream.
type StreamToSubStream struct {
	scipipe.BaseProcess
}

// NewStreamToSubStream instantiates a new StreamToSubStream process
func NewStreamToSubStream(wf *scipipe.Workflow, name string) *StreamToSubStream {
	p := &StreamToSubStream{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "substream")
	wf.AddProc(p)
	return p
}

// In returns the in-port
func (p *StreamToSubStream) In() *scipipe.InPort { return p.InPort("in") }

// OutSubStream returns the out-port
func (p *StreamToSubStream) OutSubStream() *scipipe.OutPort { return p.OutPort("substream") }

// Run runs the StreamToSubStream
func (p *StreamToSubStream) Run() {
	defer p.CloseAllOutPorts()

	scipipe.Debug.Println("Creating new information packet for the substream...")
	subStreamIP := scipipe.NewFileIP("")
	scipipe.Debug.Printf("Setting in-port of process %s to IP substream field\n", p.Name())
	subStreamIP.SubStream = p.In()

	scipipe.Debug.Printf("Sending sub-stream IP in process %s...\n", p.Name())
	p.OutSubStream().Send(subStreamIP)
	scipipe.Debug.Printf("Done sending sub-stream IP in process %s.\n", p.Name())
}
