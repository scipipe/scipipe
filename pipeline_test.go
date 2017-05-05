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

func TestRunProcessesInPipelineRunner(t *t.T) {
	proc1 := NewBogusProcess()
	proc2 := NewBogusProcess()

	pipeline := NewPipelineRunner()
	pipeline.AddProcesses(proc1, proc2)
	pipeline.Run()

	// Only the last process is supposed to be run by the pipeline directly,
	// while the others are only run if an output is pulled on an out-port,
	// but since we haven't connected the tasks here, only the last one
	// should be ran in this case.
	assert.False(t, proc1.WasRun, "Process 1 was run!")
	assert.True(t, proc2.WasRun, "Process 2 was not run!")
}

func ExamplePrintProcesses() {
	proc1 := NewBogusProcess()
	proc2 := NewBogusProcess()

	pipeline := NewPipelineRunner()
	pipeline.AddProcesses(proc1, proc2)
	pipeline.Run()

	pipeline.PrintProcesses()
	// Output:
	// Process 0: *scipipe.BogusProcess
	// Process 1: *scipipe.BogusProcess
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
