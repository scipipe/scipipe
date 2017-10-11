package scipipe

import (
	"time"
)

// Sink is a simple component that just receives InformationPacket on its In-port
// without doing anything with them
type Sink struct {
	Process
	name    string
	inPorts []*FilePort
}

// Instantiate a Sink component
func NewSink(name string) (s *Sink) {
	return &Sink{
		name:    name,
		inPorts: []*FilePort{},
	}
}

func (p *Sink) IsConnected() bool {
	return len(p.inPorts) > 0
}

func (p *Sink) Connect(outPort *FilePort) {
	newInPort := NewFilePort()
	newInPort.Connect(outPort)
	p.inPorts = append(p.inPorts, newInPort)
}

func (p *Sink) Name() string {
	return p.name
}

// Execute the Sink component
func (p *Sink) Run() {
	ok := true
	var ft *InformationPacket
	for len(p.inPorts) > 0 {
	loop:
		for i, inp := range p.inPorts {
			select {
			case ft, ok = <-inp.InChan:
				if !ok {
					Debug.Println(p.name + ": Channel closed, so deleting, and restarting loop")
					p.deleteInPortAtKey(i)
					break loop
				}
				Debug.Println(p.name+": Received file in sink: ", ft.GetPath())
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
