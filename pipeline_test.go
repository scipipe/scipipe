package scipipe

import (
	"github.com/stretchr/testify/assert"
	t "testing"
)

func testAddTasks(t *t.T) {
	task1 := NewBogusTask()
	task2 := NewBogusTask()
	pipeline := NewPipeline()
	pipeline.AddTasks(task1, task2)

	assert.NotNil(t, pipeline.tasks[0])
	assert.NotNil(t, pipeline.tasks[1])

	assert.EqualValues(t, len(pipeline.tasks), 2)

	assert.IsType(t, NewBogusTask(), pipeline.tasks[0])
	assert.IsType(t, NewBogusTask(), pipeline.tasks[1])

	pipeline.Run()
}

type BogusTask struct {
	task
}

func NewBogusTask() *BogusTask {
	return &BogusTask{}
}

func (t *BogusTask) Run() {}
