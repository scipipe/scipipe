package components

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/scipipe/scipipe"
)

// Concatenator is a process that concatenates the content of multiple files
// received in the in-port In, into one file returned on its out-port, Out.
// You can optionally specify a tag name to GroupByTag, which will make files
// go into separate output files if they have different values for that tag.
// These output files will have the tag name appended to the base file name.
type Concatenator struct {
	scipipe.BaseProcess
	OutPath    string
	GroupByTag string
}

// NewConcatenator returns a new, initialized Concatenator process
func NewConcatenator(wf *scipipe.Workflow, name string, outPath string) *Concatenator {
	p := &Concatenator{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		OutPath:     outPath,
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "out")

	wf.AddProc(p)
	return p
}

// In returns the (only) in-port for this process
func (p *Concatenator) In() *scipipe.InPort { return p.InPort("in") }

// Out returns the (only) out-port for this process
func (p *Concatenator) Out() *scipipe.OutPort { return p.OutPort("out") }

// Run runs the Concatenator process
func (p *Concatenator) Run() {
	defer p.CloseAllOutPorts()

	outIP, err := scipipe.NewFileIP(p.OutPath)
	if err != nil {
		p.Fail(err)
	}
	outFh := outIP.OpenWriteTemp()

	outIPsByTag := make(map[string]*scipipe.FileIP)
	outFhsByTag := make(map[string]*os.File)

	for inIP := range p.In().Chan {
		tagVal := inIP.Tag(p.GroupByTag)
		if tagVal != "" {
			if _, ok := outIPsByTag[tagVal]; !ok {
				newIP, err := scipipe.NewFileIP(fmt.Sprintf("%s.%s_%s", p.OutPath, p.GroupByTag, tagVal))
				if err != nil {
					p.Fail(err)
				}
				outIPsByTag[tagVal] = newIP
				outIPsByTag[tagVal].AddTag(p.GroupByTag, tagVal)
				outFhsByTag[tagVal] = outIPsByTag[tagVal].OpenWriteTemp()
			}
			dat, err := ioutil.ReadFile(inIP.Path())
			scipipe.Check(err)
			outFhsByTag[tagVal].Write(dat)
			outFhsByTag[tagVal].Write([]byte("\n"))
		} else {
			dat, err := ioutil.ReadFile(inIP.Path())
			scipipe.Check(err)
			outFh.Write(dat)
			outFh.Write([]byte("\n"))
		}
	}
	// Close file handles
	outFh.Close()
	for _, taggedFh := range outFhsByTag {
		taggedFh.Close()
	}
	// Send IPs
	p.Out().Send(outIP)
	for _, taggedIP := range outIPsByTag {
		p.Out().Send(taggedIP)
	}
}

func (p *Concatenator) Failf(msg string, parts ...interface{}) {
	p.Fail(fmt.Sprintf(msg, parts...))
}

func (p *Concatenator) Fail(msg interface{}) {
	scipipe.Failf("[Process:%s] %s", p.Name(), msg)
}
