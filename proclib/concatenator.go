package proclib

import "github.com/scipipe/scipipe"

type Concatenator struct {
	scipipe.Process
	In      *scipipe.InPort
	Out     *scipipe.OutPort
	OutPath string
}

func NewConcatenator(outPath string) *Concatenator {
	return &Concatenator{
		In:      scipipe.NewInPort(),
		Out:     scipipe.NewOutPort(),
		OutPath: outPath,
	}
}

func (proc *Concatenator) Run() {
	defer close(proc.Out.Chan)
	outFt := scipipe.NewFileTarget(proc.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range proc.In.Chan {
		fr := NewFileReader()
		go func() {
			defer close(fr.FilePath)
			fr.FilePath <- ft.GetPath()
		}()
		go fr.Run()
		for line := range fr.OutLine {
			scipipe.Debug.Println("Processing ", line, "...")
			outFh.Write(line)
		}
	}
	outFh.Close()
	outFt.Atomize()
	proc.Out.Chan <- outFt
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
