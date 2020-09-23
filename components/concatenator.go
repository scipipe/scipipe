package components

import "github.com/scipipe/scipipe"

// Concatenator is a process that concatenates the content of multiple files
// received in the in-port In, into one file returned on its out-port, Out
type Concatenator struct {
	scipipe.BaseProcess
	OutPath string
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

	outIP := scipipe.NewFileIP(p.OutPath)
	outFh := outIP.OpenWriteTemp()
	for inIP := range p.In().Chan {
		fr := NewFileToParamsReader(p.Workflow(), p.Name()+"_filereader_"+getRandString(7), inIP.Path())

		pip := scipipe.NewInParamPort(p.Name() + "temp_line_reader")
		pip.SetProcess(p)
		pip.From(fr.OutLine())

		go fr.Run()

		for line := range pip.Chan {
			outFh.WriteString(line + "\n")
		}
	}
	outFh.Close()
	p.Out().Send(outIP)
}
