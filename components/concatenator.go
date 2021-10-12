package components

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

	oipDir := filepath.Dir(outIP.Path())
	scipipe.Debug.Printf("Creating out IP dir %s\n", oipDir)
	err = os.MkdirAll(oipDir, 0777)
	if err != nil {
		p.Failf("Could not create directory: (%s) for out-IP (%s):\n%s", oipDir, outIP.Path(), err)
	}

	outFh, err := os.Create(outIP.Path())
	if err != nil {
		p.Failf("Could not open path for writing: %s\n", outIP.Path())
	}

	outIPsByTag := make(map[string]*scipipe.FileIP)
	outFhsByTag := make(map[string]*os.File)

	for inIP := range p.In().Chan {
		tagVal := inIP.Tag(p.GroupByTag)
		if tagVal != "" {
			if _, ok := outIPsByTag[tagVal]; !ok {
				outIPForTagPath := fmt.Sprintf("%s.%s_%s", p.OutPath, p.GroupByTag, tagVal)
				outIPForTag, err := scipipe.NewFileIP(outIPForTagPath)
				if err != nil {
					p.Failf("Could not create FileIP with path: %s\nOriginal error: %v", outIPForTagPath, err)
				}
				outIPForTag.AddTag(p.GroupByTag, tagVal)
				outIPsByTag[tagVal] = outIPForTag
				outFh, err := os.Create(outIPForTag.Path())
				if err != nil {
					p.Failf("Could not create path: %s\nOriginal error: %v", outIPForTag.Path(), err)
				}
				outFhsByTag[tagVal] = outFh
			}
			dat, err := ioutil.ReadFile(inIP.Path())
			if err != nil {
				p.Failf("Could not read file: %s\n", inIP.Path())
			}
			outFhsByTag[tagVal].Write(append(dat))
			if err != nil {
				p.Failf("Could not write to file: %s\n", outIPsByTag[tagVal].Path())
			}
			outFhsByTag[tagVal].Write(append([]byte("\n")))
			if err != nil {
				p.Failf("Could not write to file: %s\n", outIPsByTag[tagVal].Path())
			}
		} else {
			dat, err := ioutil.ReadFile(inIP.Path())
			if err != nil {
				p.Failf("Could not read file: %s\n", inIP.Path())
			}
			_, err = outFh.Write(append(dat))
			if err != nil {
				p.Failf("Could not write to file: %s\n", outIP.Path())
			}
			_, err = outFh.Write(append([]byte("\n")))
			if err != nil {
				p.Failf("Could not write to file: %s\n", outIP.Path())
			}
		}
	}

	// Close file handles
	err = outFh.Close()
	if err != nil {
		p.Failf("Could not close file handle: %s\n", outIP.Path())
	}
	for _, taggedFh := range outFhsByTag {
		taggedFh.Close()
	}

	// Send IPs
	p.Out().Send(outIP)
	for _, taggedIP := range outIPsByTag {
		p.Out().Send(taggedIP)
	}
}
