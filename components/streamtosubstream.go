package components

import "github.com/scipipe/scipipe"

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
func NewStreamToSubStream(name string) *StreamToSubStream {
	return &StreamToSubStream{
		name:         name,
		In:           scipipe.NewFilePort(),
		OutSubStream: scipipe.NewFilePort(),
	}
}

// Run the StreamToSubStream
func (p *StreamToSubStream) Run() {
	defer p.OutSubStream.Close()

	subStreamIP := scipipe.NewInformationPacket("")
	subStreamIP.SubStream = p.In

	p.OutSubStream.Chan <- subStreamIP
}

func (p *StreamToSubStream) Name() string {
	return p.name
}

func (p *StreamToSubStream) IsConnected() bool {
	return p.In.IsConnected() && p.OutSubStream.IsConnected()
}
