package scipipe

type MultiplexerX2 struct {
	Process
	InFile   chan *FileTarget
	OutFile1 chan *FileTarget
	OutFile2 chan *FileTarget
}

func NewMultiplexerX2() *MultiplexerX2 {
	return &MultiplexerX2{
		InFile:   make(chan *FileTarget, BUFSIZE),
		OutFile1: make(chan *FileTarget, BUFSIZE),
		OutFile2: make(chan *FileTarget, BUFSIZE),
	}
}

func (proc *MultiplexerX2) Run() {
	defer close(proc.OutFile1)
	defer close(proc.OutFile2)

	for ft := range proc.InFile {
		proc.OutFile1 <- ft
		proc.OutFile2 <- ft
	}
}
