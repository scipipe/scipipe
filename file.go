package scipipe

import (
	"os"
	"time"
)

type fileTarget struct {
	path string
}

func NewFileTarget(path string) *fileTarget {
	ft := new(fileTarget)
	ft.path = path
	return ft
}

func (ft *fileTarget) GetPath() string {
	return ft.path
}

func (ft *fileTarget) GetTempPath() string {
	return ft.path + ".tmp"
}

func (ft *fileTarget) Atomize() {
	time.Sleep(1 * time.Second) // TODO: Remove in production. Just for demo purposes!
	err := os.Rename(ft.GetTempPath(), ft.path)
	Check(err)
}
