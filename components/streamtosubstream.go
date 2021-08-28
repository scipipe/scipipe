package components

import (
	"fmt"
	"io/ioutil"
	"os"

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

	// create a temporary file, with a _scipipe prefix
	tmpfile, err := ioutil.TempFile("", "_scipipe_tmp.")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	scipipe.Debug.Println("Creating new information packet for the substream...")
	subStreamIP, err := scipipe.NewFileIP(tmpfile.Name())
	if err != nil {
		p.Fail(err)
	}
	scipipe.Debug.Printf("Setting in-port of process %s to IP substream field\n", p.Name())
	subStreamIP.SubStream = p.In()

	scipipe.Debug.Printf("Sending sub-stream IP in process %s...\n", p.Name())
	p.OutSubStream().Send(subStreamIP)
	scipipe.Debug.Printf("Done sending sub-stream IP in process %s.\n", p.Name())
}

func (p *StreamToSubStream) Failf(msg string, parts ...interface{}) {
	p.Fail(fmt.Sprintf(msg, parts...))
}

func (p *StreamToSubStream) Fail(msg interface{}) {
	scipipe.Failf("[Process:%s] %s", p.Name(), msg)
}
