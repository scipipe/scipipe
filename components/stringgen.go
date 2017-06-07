package components

import "github.com/scipipe/scipipe"

// StringGen takes a number of strings and returns a generator process
// which sends the strings, one by one, on its `Out` port
type StringGen struct {
	scipipe.Process
	Strings []string
	Out     *scipipe.ParamPort
}

// Instantiate a new StringGen
func NewStringGen(strings ...string) *StringGen {
	return &StringGen{
		Out:     scipipe.NewParamPort(),
		Strings: strings,
	}
}

// Run the StringGen
func (proc *StringGen) Run() {
	defer proc.Out.Close()
	for _, str := range proc.Strings {
		proc.Out.Chan <- str
	}
}

func (proc *StringGen) IsConnected() bool {
	return proc.Out.IsConnected()
}
