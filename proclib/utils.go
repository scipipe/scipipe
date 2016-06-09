package proclib

import (
	"github.com/scipipe/scipipe"
	"os"
)

func Check(err error) {
	if err != nil {
		scipipe.Error.Println(err.Error())
		os.Exit(1)
	}
}
