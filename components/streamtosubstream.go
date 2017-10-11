package components

import (
	"github.com/scipipe/scipipe"
)

// StreamToSubStream takes a normal stream of InformationPacket's representing
// individual files, and returns one InformationPacket where the incoming IPs
// are sent on its substream.
type StreamToSubStream struct {
	scipipe.Process
	name         string
	In           *scipipe.FilePort
	OutSubStream *scipipe.FilePort
}

// Instantiate a new StreamToSubStream
func NewStreamToSubStream(wf *scipipe.Workflow, name string) *StreamToSubStream {
	stss := &StreamToSubStream{
		name:         name,
		In:           scipipe.NewFilePort(),
		OutSubStream: scipipe.NewFilePort(),
	}
	wf.AddProc(stss)
	return stss
}

// Run the StreamToSubStream
func (p *StreamToSubStream) Run() {
	defer p.OutSubStream.Close()
	scipipe.Debug.Println("Creating new information packet for the substream...")
	subStreamIP := scipipe.NewInformationPacket("")
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
