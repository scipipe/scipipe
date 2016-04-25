package scipipe

// Duplicator duplicates *FileTarget received on
// the InFile in-bound port, and sends them on the out-ports
// OutFile1 and OutFile2
type Duplicator struct {
	Process
	InFile   chan *FileTarget
	OutFile1 chan *FileTarget
	OutFile2 chan *FileTarget
}

// NewDuplicator creates a new Duplicator process
func NewDuplicator() *Duplicator {
	return &Duplicator{
		InFile:   make(chan *FileTarget, BUFSIZE),
		OutFile1: make(chan *FileTarget, BUFSIZE),
		OutFile2: make(chan *FileTarget, BUFSIZE),
	}
}

// Run runs the Duplicator process
func (proc *Duplicator) Run() {
	defer close(proc.OutFile1)
	defer close(proc.OutFile2)

	for ft := range proc.InFile {
		proc.OutFile1 <- ft
		proc.OutFile2 <- ft
	}
}
