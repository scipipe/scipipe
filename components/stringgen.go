package components

import "github.com/scipipe/scipipe"

// StringGen takes a number of strings and returns a generator process
// which sends the strings, one by one, on its `Out` port
type StringGen struct {
	scipipe.EmptyWorkflowProcess
	name    string
	Strings []string
	Out     *scipipe.ParamOutPort
}

// NewStringGen instantiate a new StringGen
func NewStringGen(wf *scipipe.Workflow, name string, strings ...string) *StringGen {
	sg := &StringGen{
		name:    name,
		Out:     scipipe.NewParamOutPort(),
		Strings: strings,
	}
	wf.AddProc(sg)
	return sg
}

// Name returns the name of the StringGen process
func (proc *StringGen) Name() string {
	return proc.name
}

// OutParamPorts returns the out-param-ports of the StringGen process
func (proc *StringGen) OutParamPorts() map[string]*scipipe.ParamOutPort {
	return map[string]*scipipe.ParamOutPort{"out": proc.Out}
}

// Run runs the StringGen
func (proc *StringGen) Run() {
	defer proc.Out.Close()
	for _, str := range proc.Strings {
		proc.Out.Send(str)
	}
}

// IsConnected tells whether all ports of the StringGen process are connected
func (proc *StringGen) IsConnected() bool {
	return proc.Out.IsConnected()
}
