package scipipe

type strToByte struct {
	task
	In  chan string
	Out chan []byte
}

func NewStrToByte(pl *Pipeline) *strToByte {
	t := &strToByte{
		In:  make(chan string, BUFSIZE),
		Out: make(chan []byte, BUFSIZE),
	}
	pl.AddTask(t)
	return t
}

func (proc *strToByte) Run() {
	defer close(proc.Out)
	for line := range proc.In {
		proc.Out <- []byte(line)
	}
}
