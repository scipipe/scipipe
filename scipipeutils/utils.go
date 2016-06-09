package scipipeutils

import (
	"os"
)

func Check(err error) {
	if err != nil {
		Error.Println(err.Error())
		os.Exit(1)
	}
}
