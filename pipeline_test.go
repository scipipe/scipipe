package scipipe

import (
	"github.com/stretchr/testify/assert"
	t "testing"
)

func testAddProcesses(t *t.T) {
	proc1 := NewBogusProcess()
	proc2 := NewBogusProcess()
	pipeline := NewPipeline()
	pipeline.AddProcesses(proc1, proc2)

	assert.NotNil(t, pipeline.processes[0])
	assert.NotNil(t, pipeline.processes[1])

	assert.EqualValues(t, len(pipeline.processes), 2)

	assert.IsType(t, NewBogusProcess(), pipeline.processes[0])
	assert.IsType(t, NewBogusProcess(), pipeline.processes[1])

	pipeline.Run()
}

type BogusProcess struct {
	process
}

func NewBogusProcess() *BogusProcess {
	return &BogusProcess{}
}

func (t *BogusProcess) Run() {}
