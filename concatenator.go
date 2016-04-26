package scipipe

type Concatenator struct {
	Process
	In      chan *FileTarget
	Out     chan *FileTarget
	OutPath string
}

func NewConcatenator(outPath string) *Concatenator {
	return &Concatenator{
		In:      make(chan *FileTarget, BUFSIZE),
		Out:     make(chan *FileTarget, BUFSIZE),
		OutPath: outPath,
	}
}

func (proc *Concatenator) Run() {
	defer close(proc.Out)
	outFt := NewFileTarget(proc.OutPath)
	outFh := outFt.OpenWriteTemp()
	for ft := range proc.In {
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
	proc.Out <- outFt
}
