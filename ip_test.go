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
	assertPathsEqual(t, ip.Path(), TESTPATH)
	assertPathsEqual(t, ip.TempPath(), TESTPATH+".tmp")
	assertPathsEqual(t, ip.FifoPath(), TESTPATH+".fifo")
}

func assertPathsEqual(t *testing.T, path1 string, path2 string) {
	assert.Equal(t, path1, path2, "Wrong path returned! (Was", path1, "but should be", path2, ")")
}
