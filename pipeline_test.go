package scipipe

import (
	"github.com/stretchr/testify/assert"
	t "testing"
)

func TestAddProcesses(t *t.T) {
	proc1 := NewBogusProcess()
	proc2 := NewBogusProcess()
	pipeline := NewPipeline()
	pipeline.AddProcesses(proc1, proc2)

	assert.EqualValues(t, len(pipeline.processes), 2)

	assert.IsType(t, &BogusProcess{}, pipeline.processes[0])
	assert.IsType(t, &BogusProcess{}, pipeline.processes[1])
}

type BogusProcess struct {
	process
}

func NewBogusProcess() *BogusProcess {
	return &BogusProcess{}
}

func (t *BogusProcess) Run() {}
