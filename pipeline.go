package scipipe

import (
	"fmt"
	"reflect"
)

type Pipeline struct {
	tasks []task
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

func (pl *Pipeline) AddTask(t task) {
	pl.tasks = append(pl.tasks, t)
}

func (pl *Pipeline) AddTasks(tasks ...task) {
	for _, task := range tasks {
		pl.AddTask(task)
	}
}

func (pl *Pipeline) PrintTasks() {
	for i, task := range pl.tasks {
		fmt.Printf("Task %d: %v\n", i, reflect.TypeOf(task))
	}
}

func (pl *Pipeline) Run() {
	for i, task := range pl.tasks {
		if i < len(pl.tasks)-1 {
			go task.Run()
		} else {
			task.Run()
		}
	}
}
