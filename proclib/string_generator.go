package proclib

import "github.com/scipipe/scipipe"

// StringGenerator takes a number of strings and returns a generator process
// which sends the strings, one by one, on its `Out` port
type StringGenerator struct {
	scipipe.Process
	Strings []string
	Out     *scipipe.ParamPort
}

// Instantiate a new StringGenerator
func NewStringGenerator(strings ...string) *StringGenerator {
	return &StringGenerator{
		Out:     scipipe.NewParamPort(),
		Strings: strings,
	}
}

// Run the StringGenerator
func (proc *StringGenerator) Run() {
	defer proc.Out.Close()
	for _, str := range proc.Strings {
		proc.Out.Chan <- str
	}
}

func (proc *StringGenerator) IsConnected() bool {
	return proc.Out.IsConnected()
}
