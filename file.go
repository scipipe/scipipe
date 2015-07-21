package scipipe

import (
	"io/ioutil"
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

func (ft *FileTarget) Open() *os.File {
	f, err := os.Open(ft.GetPath())
	Check(err)
	return f
}

func (ft *FileTarget) Read() []byte {
	dat, err := ioutil.ReadFile(ft.GetPath())
	Check(err)
	return dat
}

func (ft *FileTarget) Write(dat []byte) {
	err := ioutil.WriteFile(ft.GetTempPath(), dat, 0644)
	ft.Atomize()
	Check(err)
}

func (ft *FileTarget) Atomize() {
	time.Sleep(1 * time.Millisecond) // TODO: Remove in production. Just for demo purposes!
	err := os.Rename(ft.GetTempPath(), ft.path)
	Check(err)
}

func (ft *FileTarget) Exists() bool {
	if _, err := os.Stat(ft.GetPath()); err == nil {
		return true
	}
	return false
}
