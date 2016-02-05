package scipipe

// Base interface for all processes
type process interface {
	Run()
}

type BaseProcess struct {
	process
}

func NewBaseProcess() *BaseProcess {
	return &BaseProcess{}
}
