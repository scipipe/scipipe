package components

import "github.com/scipipe/scipipe"

// StreamToSubStream takes a normal stream of InformationPacket's representing
// individual files, and returns one InformationPacket where the incoming IPs
// are sent on its substream.
type StreamToSubStream struct {
	scipipe.Process
	In           *scipipe.FilePort
	OutSubStream *scipipe.FilePort
}

// Instantiate a new StreamToSubStream
func NewStreamToSubStream() *StreamToSubStream {
	return &StreamToSubStream{}
}

// Run the StreamToSubStream
func (proc *StreamToSubStream) Run() {
	defer proc.OutSubStream.Close()

	subStreamIP := scipipe.NewInformationPacket("")
	scipipe.Connect(proc.In, subStreamIP.SubStream)

	proc.OutSubStream.Chan <- subStreamIP
}

func (proc *StreamToSubStream) IsConnected() bool {
	return proc.In.IsConnected() && proc.OutSubStream.IsConnected()
}
