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
func (proc *Concatenator) Name() string {
	return proc.name
}

// Run runs the Concatenator process
func (proc *Concatenator) Run() {
	defer proc.Out.Close()

	outFt := scipipe.NewIP(proc.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range proc.In.Chan {
		fr := NewFileReader(proc.workflow, proc.Name()+"_filereader")
		go func() {
			defer close(fr.FilePath)
			fr.FilePath <- ft.GetPath()
		}()
		for line := range fr.OutLine {
			scipipe.Debug.Println("Processing ", line, "...")
			outFh.Write([]byte(line))
		}
	}
	outFh.Close()
	outFt.Atomize()
	proc.Out.Send(outFt)
}

// IsConnected tells whether all ports of the Concatenator process are connected
func (proc *Concatenator) IsConnected() bool {
	isConnected := true
	if !proc.In.IsConnected() {
		scipipe.Error.Println("Concatenator: Port 'In' is not connected!")
		isConnected = false
	}
	if !proc.Out.IsConnected() {
		scipipe.Error.Println("Concatenator: Port 'Out' is not connected!")
		isConnected = false
	}
	return isConnected
}
