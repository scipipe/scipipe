package components

import "github.com/scipipe/scipipe"

type Concatenator struct {
	scipipe.Process
	name    string
	In      *scipipe.FilePort
	Out     *scipipe.FilePort
	OutPath string
}

func NewConcatenator(name string, outPath string) *Concatenator {
	return &Concatenator{
		name:    name,
		In:      scipipe.NewFilePort(),
		Out:     scipipe.NewFilePort(),
		OutPath: outPath,
	}
}

func (proc *Concatenator) Name() string {
	return proc.name
}

func (proc *Concatenator) Run() {
	defer close(proc.Out.InChan)
	outFt := scipipe.NewInformationPacket(proc.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range proc.In.InChan {
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
