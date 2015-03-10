package scipipe

type Task struct {
	InPorts      map[string]chan *FileTarget
	OutPorts     map[string]chan *FileTarget
	InPaths      map[string]string
	OutPathFuncs map[string]func() string
}
