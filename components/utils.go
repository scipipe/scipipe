package components

import (
	"github.com/scipipe/scipipe"
)

func Check(err error) {
	if err != nil {
		scipipe.Error.Fatalln(err.Error())
	}
}
