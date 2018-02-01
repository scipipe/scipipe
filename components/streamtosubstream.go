package components

import (
	"github.com/scipipe/scipipe"
)

// StreamToSubStream takes a normal stream of IP's representing
// individual files, and returns one IP where the incoming IPs
// are sent on its substream.
type StreamToSubStream struct {
	scipipe.Process
	name         string
	In           *scipipe.Port
	OutSubStream *scipipe.Port
}

// Instantiate a new StreamToSubStream
func NewStreamToSubStream(wf *scipipe.Workflow, name string) *StreamToSubStream {
	stss := &StreamToSubStream{
		name:         name,
		In:           scipipe.NewPort(),
		OutSubStream: scipipe.NewPort(),
	}
	wf.AddProc(stss)
	return stss
}

// Run the StreamToSubStream
func (p *StreamToSubStream) Run() {
	defer p.OutSubStream.Close()
	go p.In.RunMergeInputs()

	scipipe.Debug.Println("Creating new information packet for the substream...")
	subStreamIP := scipipe.NewIP("")
	scipipe.Debug.Printf("Setting in-port of process %s to IP substream field\n", p.Name())
	subStreamIP.SubStream = p.In

	scipipe.Debug.Printf("Sending sub-stream IP in process %s...\n", p.Name())
	p.OutSubStream.Send(subStreamIP)
	scipipe.Debug.Printf("Done sending sub-stream IP in process %s.\n", p.Name())
}

func (p *StreamToSubStream) Name() string {
	return p.name
}

func (p *StreamToSubStream) IsConnected() bool {
	return p.In.IsConnected() && p.OutSubStream.IsConnected()
}
