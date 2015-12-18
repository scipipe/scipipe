package scipipe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TESTPATH = "somepath.txt"
)

func TestFileTargetPaths(t *testing.T) {
	ft := NewFileTarget(TESTPATH)
	assertPathsEqual(t, ft.GetPath(), TESTPATH)
	assertPathsEqual(t, ft.GetTempPath(), TESTPATH+".tmp")
	assertPathsEqual(t, ft.GetFifoPath(), TESTPATH+".fifo")
}

func assertPathsEqual(t *testing.T, path1 string, path2 string) {
	assert.Equal(t, path1, path2, "Wrong path returned! (Was", path1, "but should be", path2, ")")
}
