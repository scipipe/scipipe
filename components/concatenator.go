package components

import "github.com/scipipe/scipipe"

type Concatenator struct {
	name     string
	In       *scipipe.Port
	Out      *scipipe.Port
	OutPath  string
	workflow *scipipe.Workflow
}

func NewConcatenator(wf *scipipe.Workflow, name string, outPath string) *Concatenator {
	concat := &Concatenator{
		name:     name,
		In:       scipipe.NewPort(),
		Out:      scipipe.NewPort(),
		OutPath:  outPath,
		workflow: wf,
	}
	wf.AddProc(concat)
	return concat
}

func (proc *Concatenator) Name() string {
	return proc.name
}

func (proc *Concatenator) Run() {
	defer proc.Out.Close()
	go proc.In.RunMergeInputs()

	outFt := scipipe.NewIP(proc.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range proc.In.InChan {
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
