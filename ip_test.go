package scipipe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TESTPATH = "somepath.txt"
)

func TestIPPaths(t *testing.T) {
	ip := NewIP(TESTPATH)
	assertPathsEqual(t, ip.GetPath(), TESTPATH)
	assertPathsEqual(t, ip.GetTempPath(), TESTPATH+".tmp")
	assertPathsEqual(t, ip.GetFifoPath(), TESTPATH+".fifo")
}

func assertPathsEqual(t *testing.T, path1 string, path2 string) {
	assert.Equal(t, path1, path2, "Wrong path returned! (Was", path1, "but should be", path2, ")")
}
