package scipipe

type Concatenator struct {
	Process
	In      *InPort
	Out     *OutPort
	OutPath string
}

func NewConcatenator(outPath string) *Concatenator {
	return &Concatenator{
		In:      NewInPort(),
		Out:     NewOutPort(),
		OutPath: outPath,
	}
}

func (proc *Concatenator) Run() {
	defer close(proc.Out.Chan)
	outFt := NewFileTarget(proc.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range proc.In.Chan {
		fr := NewFileReader()
		go func() {
			defer close(fr.FilePath)
			fr.FilePath <- ft.GetPath()
		}()
		go fr.Run()
		for line := range fr.OutLine {
			Debug.Println("Processing ", line, "...")
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
		Error.Println("Concatenator: Port 'In' is not connected!")
		isConnected = false
	}
	if !proc.Out.IsConnected() {
		Error.Println("Concatenator: Port 'Out' is not connected!")
		isConnected = false
	}
	return isConnected
}
