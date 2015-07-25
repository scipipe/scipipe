package scipipe

type process interface {
	Run()
}

type BaseProcess struct {
	process
}

func NewBaseProcess() *BaseProcess {
	return &BaseProcess{}
}
