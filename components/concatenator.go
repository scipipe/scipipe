package components

import "github.com/scipipe/scipipe"

type Concatenator struct {
	scipipe.Process
	In      *scipipe.FilePort
	Out     *scipipe.FilePort
	OutPath string
}

func NewConcatenator(outPath string) *Concatenator {
	return &Concatenator{
		In:      scipipe.NewFilePort(),
		Out:     scipipe.NewFilePort(),
		OutPath: outPath,
	}
}

func (proc *Concatenator) Run() {
	defer close(proc.Out.Chan)
	outFt := scipipe.NewInformationPacket(proc.OutPath)
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
			outFh.Write([]byte(line))
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
