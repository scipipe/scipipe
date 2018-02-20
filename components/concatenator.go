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

	outFt := scipipe.NewFileIP(p.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range p.In().Chan {

		fr := NewFileReader(p.Workflow(), p.Name()+"_filereader_"+getRandString(7))
		pop := scipipe.NewParamOutPort("temp_filepath_feeder")
		pop.SetProcess(p)
		fr.InFilePath().Connect(pop)
		go func() {
			defer pop.Close()
			pop.Send(ft.Path())
		}()

		pip := scipipe.NewParamInPort(p.Name() + "temp_line_reader")
		pip.SetProcess(p)
		pip.Connect(fr.OutLine())

		go fr.Run()
		for line := range pip.Chan {
			scipipe.Debug.Println("Processing ", line, "...")
			outFh.Write([]byte(line))
		}
	}
	outFh.Close()
	outFt.Atomize()
	p.Out().Send(outFt)
}
