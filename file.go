package scipipe

import (
	"os"
	"time"
)

type FileTarget struct {
	path string
}

func NewFileTarget(path string) *FileTarget {
	ft := new(FileTarget)
	ft.path = path
	return ft
}

func (ft *FileTarget) GetPath() string {
	return ft.path
}

func (ft *FileTarget) GetTempPath() string {
	return ft.path + ".tmp"
}

func (ft *FileTarget) Atomize() {
	time.Sleep(1 * time.Second) // TODO: Remove in production. Just for demo purposes!
	err := os.Rename(ft.GetTempPath(), ft.path)
	Check(err)
}

func (ft *FileTarget) Exists() bool {
	if _, err := os.Stat(ft.GetPath()); err == nil {
		return true
	}
	return false
}
