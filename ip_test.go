package scipipe

import (
	"testing"
)

const (
	TESTPATH = "somepath.txt"
)

func TestIPPaths(t *testing.T) {
	ip, err := NewFileIP(TESTPATH)
	Check(err)
	assertPathsEqual(t, ip.Path(), TESTPATH)
	assertPathsEqual(t, ip.TempPath(), TESTPATH)
	assertPathsEqual(t, ip.FifoPath(), TESTPATH+".fifo")
}

func assertPathsEqual(t *testing.T, path1 string, path2 string) {
	if path1 != path2 {
		t.Errorf("Wrong path returned. Was %s but should be %s\n", path1, path2)
	}
}

func TestPathIsValid(t *testing.T) {
	for _, tc := range []struct {
		path      string
		wantValid bool
	}{
		{path: "filename.txt", wantValid: true},
		{path: "file name.txt", wantValid: false},
		{path: "./directory/long-file_name.txt", wantValid: true},
		{path: `\\Server\path`, wantValid: false},
	} {
		haveValid, err := pathIsValid(tc.path)
		Check(err)
		if haveValid != tc.wantValid {
			t.Fatalf("Valid status for path '%s' was %v, not %v as expected", tc.path, haveValid, tc.wantValid)
		}
	}
}
