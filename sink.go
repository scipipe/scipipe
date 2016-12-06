package scipipe

import (
	"time"
)

// Sink is a simple component that just receives FileTargets on its In-port
// without doing anything with them
type Sink struct {
	Process
	inPorts []*FilePort
}

// Instantiate a Sink component
func NewSink() (s *Sink) {
	return &Sink{
		inPorts: []*FilePort{},
	}
}

func (proc *Sink) IsConnected() bool {
	if len(proc.inPorts) > 0 {
		return true
	} else {
		return false
	}
}

func (proc *Sink) Connect(outPort *FilePort) {
	newInPort := NewFilePort()
	newInPort.Connect(outPort)
	proc.inPorts = append(proc.inPorts, newInPort)
}

// Execute the Sink component
func (proc *Sink) Run() {
	ok := true
	var ft *FileTarget
	for len(proc.inPorts) > 0 {
		for i, inp := range proc.inPorts {
			select {
			case ft, ok = <-inp.Chan:
				if !ok {
					proc.deleteInPortAtKey(i)
					continue
				}
				Debug.Println("Received file in sink: ", ft.GetPath())
			default:
				Debug.Printf("No receive on inport %d, so continuing ...\n", i)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
