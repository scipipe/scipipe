package scipipe

import (
	"testing"
)

const (
	TESTPATH = "somepath.txt"
)

func TestIPPaths(t *testing.T) {
	ip := NewFileIP(TESTPATH)
	assertPathsEqual(t, ip.Path(), TESTPATH)
	assertPathsEqual(t, ip.TempPath(), TESTPATH+".tmp")
	assertPathsEqual(t, ip.FifoPath(), TESTPATH+".fifo")
}

func assertPathsEqual(t *testing.T, path1 string, path2 string) {
	if path1 != path2 {
		t.Errorf("Wrong path returned. Was %s but should be %s\n", path1, path2)
	}
}
