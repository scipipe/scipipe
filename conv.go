package scipipe

// SciPipe component that converts packets of string type to byte
type strToByte struct {
	process
	In  chan string
	Out chan []byte
}

func NewStrToByte() *strToByte {
	return &strToByte{
		In:  make(chan string, BUFSIZE),
		Out: make(chan []byte, BUFSIZE),
	}
}

func (proc *strToByte) Run() {
	defer close(proc.Out)
	for line := range proc.In {
		proc.Out <- []byte(line)
	}
}
