package scipipe

import (
	"github.com/stretchr/testify/assert"
	t "testing"
)

func testAddTasks(t *t.T) {
	task1 := NewBaseTask()
	task2 := NewBaseTask()
	pipeline := NewPipeline()
	pipeline.AddTasks(task1, task2)
	assert.NotNil(t, pipeline.tasks[0])
	assert.NotNil(t, pipeline.tasks[1])
	assert.IsType(t, NewBaseTask(), pipeline.tasks[0])
	assert.IsType(t, NewBaseTask(), pipeline.tasks[1])
}
