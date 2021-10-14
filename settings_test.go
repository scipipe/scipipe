package scipipe

import (
	"os"
	"testing"
)

func TestGetBufsize(t *testing.T) {
	initTestLogs()

	wantBufSize := 1234
	err := os.Setenv("SCIPIPE_BUFSIZE", "1234")
	Check(err)
	haveBufSize := getBufsize()

	if haveBufSize != wantBufSize {
		t.Errorf("Got wrong buffert size from getBufsize(): %d, wanted: %d\n", haveBufSize, wantBufSize)
	}
}
