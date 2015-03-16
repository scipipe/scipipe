package scipipe

import (
	"testing"
)

const (
	TESTPATH = "somepath.txt"
)

func TestFileTargetPath(t *testing.T) {
	ft := NewFileTarget(TESTPATH)
	path := ft.GetPath()
	if path != TESTPATH {
		t.Errorf("Path not properly initialized! (Was", path, "but should be", TESTPATH)
	}
}
