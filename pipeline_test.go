package scipipe

import (
	"github.com/stretchr/testify/assert"
	"sync"
	t "testing"
)

func TestAddProcesses(t *t.T) {
	InitLogError()

	proc1 := NewBogusProcess()
	proc2 := NewBogusProcess()
	pipeline := NewPipelineRunner()
	pipeline.AddProcesses(proc1, proc2)

	assert.EqualValues(t, len(pipeline.processes), 2)

	assert.IsType(t, &BogusProcess{}, pipeline.processes[0], "Process 1 was not of the right type!")
	assert.IsType(t, &BogusProcess{}, pipeline.processes[1], "Process 2 was not of the right type!")
}

// --------------------------------
// Helper stuff
// --------------------------------

// A process with does just satisfy the Process interface, without doing any
// actual work.
type BogusProcess struct {
	Process
	WasRun     bool
	WasRunLock sync.Mutex
}

var bogusWg sync.WaitGroup

func NewBogusProcess() *BogusProcess {
	return &BogusProcess{WasRun: false}
}

func (p *BogusProcess) Run() {
	p.WasRunLock.Lock()
	p.WasRun = true
	p.WasRunLock.Unlock()
}

func (p *BogusProcess) IsConnected() bool {
	return true
}
