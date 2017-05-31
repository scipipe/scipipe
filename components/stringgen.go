package components

import "github.com/scipipe/scipipe"

// StringGen takes a number of strings and returns a generator process
// which sends the strings, one by one, on its `Out` port
type StringGen struct {
	scipipe.Process
	name    string
	Strings []string
	Out     *scipipe.ParamPort
}

// Instantiate a new StringGen
func NewStringGen(name string, strings ...string) *StringGen {
	return &StringGen{
		name:    name,
		Out:     scipipe.NewParamPort(),
		Strings: strings,
	}
}

func (proc *StringGen) Name() string {
	return proc.name
}

// Run the StringGen
func (proc *StringGen) Run() {
	defer proc.Out.Close()
	for _, str := range proc.Strings {
		proc.Out.Send(str)
	}
}

func (proc *StringGen) IsConnected() bool {
	return proc.Out.IsConnected()
}
