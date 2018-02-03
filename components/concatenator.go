package components

import "github.com/scipipe/scipipe"

// Concatenator is a process that concatenates the content of multiple files
// received in the in-port In, into one file returned on its out-port, Out
type Concatenator struct {
	scipipe.EmptyWorkflowProcess
	name     string
	In       *scipipe.InPort
	Out      *scipipe.OutPort
	OutPath  string
	workflow *scipipe.Workflow
}

// NewConcatenator returns a new, initialized Concatenator process
func NewConcatenator(wf *scipipe.Workflow, name string, outPath string) *Concatenator {
	concat := &Concatenator{
		name:     name,
		In:       scipipe.NewInPort("in"),
		Out:      scipipe.NewOutPort("out"),
		OutPath:  outPath,
		workflow: wf,
	}
	wf.AddProc(concat)
	return concat
}

// Name returns the name of the Concatenator process
func (p *Concatenator) Name() string {
	return p.name
}

// InPorts returns all the in-ports for the process
func (p *Concatenator) InPorts() map[string]*scipipe.InPort {
	return map[string]*scipipe.InPort{
		p.In.Name(): p.In,
	}
}

// OutPorts returns all the out-ports for the process
func (p *Concatenator) OutPorts() map[string]*scipipe.OutPort {
	return map[string]*scipipe.OutPort{
		p.Out.Name(): p.Out,
	}
}

// Run runs the Concatenator process
func (p *Concatenator) Run() {
	defer p.Out.Close()

	outFt := scipipe.NewIP(p.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range p.In.Chan {

		fr := NewFileReader(p.workflow, p.Name()+"_filereader")
		pop := scipipe.NewParamOutPort("temp_filepath_feeder")
		pop.Process = p
		fr.FilePath.Connect(pop)
		go func() {
			defer pop.Close()
			pop.Send(ft.Path())
		}()

		pip := scipipe.NewParamInPort(p.Name() + "temp_line_reader")
		pip.Process = p
		pip.Connect(fr.OutLine)

		go fr.Run()
		for line := range pip.Chan {
			scipipe.Debug.Println("Processing ", line, "...")
			outFh.Write([]byte(line))
		}
	}
	outFh.Close()
	outFt.Atomize()
	p.Out.Send(outFt)
}

// IsConnected tells whether all ports of the Concatenator process are connected
func (p *Concatenator) IsConnected() bool {
	isConnected := true
	if !p.In.IsConnected() {
		scipipe.Error.Println("Concatenator: Port 'In' is not connected!")
		isConnected = false
	}
	if !p.Out.IsConnected() {
		scipipe.Error.Println("Concatenator: Port 'Out' is not connected!")
		isConnected = false
	}
	return isConnected
}
