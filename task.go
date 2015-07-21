package scipipe

type task interface {
	Run()
}

type BaseTask struct {
	task
}

func NewBaseTask() *BaseTask {
	return &BaseTask{}
}
