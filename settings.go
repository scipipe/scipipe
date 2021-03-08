package scipipe

import (
	"os"
	"strconv"
)

var (
	BUFSIZE = 128
)

func getBufsize() int {
	// BUFSIZE is the standard buffer size used for channels connecting processes
	if bufSizeStr, envSet := os.LookupEnv("SCIPIPE_BUFSIZE"); envSet {
		bufSize, err := strconv.Atoi(bufSizeStr)
		if err != nil {
			Failf("Could not convert value of SCIPIPE_BUFSIZE to integer: %s\n", bufSizeStr)
		}
		return bufSize
	}
	return BUFSIZE
}
