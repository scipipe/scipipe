package components

import (
	"github.com/scipipe/scipipe"
	"strings"
)

// FileReader takes a file path on its FilePath in-port, and returns the file
// content as []byte on its out-port Out
type IpToParamConverter struct {
	scipipe.Process
	name     string
	InFile   *scipipe.FilePort
	OutParam *scipipe.ParamPort
}

// Instantiate a new IpToParamConverter
func NewIpToParamConverter(wf *scipipe.Workflow, name string) *IpToParamConverter {
	p := &IpToParamConverter{
		name:     name,
		InFile:   scipipe.NewFilePort(),
		OutParam: scipipe.NewParamPort(),
	}
	wf.AddProc(p)
	return p
}

func (p *IpToParamConverter) Name() string {
	return p.name
}

func (p *IpToParamConverter) IsConnected() bool {
	return p.InFile.IsConnected() && p.OutParam.IsConnected()
}

// Run the IpToParamConverter
func (p *IpToParamConverter) Run() {
	defer p.OutParam.Close()
	go p.InFile.RunMergeInputs()

	for ip := range p.InFile.InChan {
		s := string(ip.Read())
		s = strings.Trim(s, " \r\n\t")
		p.OutParam.Send(s)
	}
}
