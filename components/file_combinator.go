package components

import (
	"github.com/scipipe/scipipe"
)

type FileCombinator struct {
	scipipe.BaseProcess
	globPatterns []string
}

// NewFileCombinator returns a new initialized FileCombinator process
func NewFileCombinator(wf *scipipe.Workflow, name string) *FileCombinator {
	p := &FileCombinator{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
	}
	wf.AddProc(p)
	return p
}

// Out returns the outport
func (p *FileCombinator) Out(pName string) *scipipe.OutPort { return p.OutPort(pName) }

// In returns the in-port
func (p *FileCombinator) In(pName string) *scipipe.InPort {
	if _, ok := p.InPorts()[pName]; !ok {
		p.InitInPort(p, pName)
	}
	// Initialize a corresponding outport with the same name, for each inport
	if _, ok := p.OutPorts()[pName]; !ok {
		p.InitOutPort(p, pName)
	}
	return p.InPort(pName)
}

// Run runs the FileCombinator process
func (p *FileCombinator) Run() {
	defer p.CloseAllOutPorts()

	inIPs := map[string][]*scipipe.FileIP{}

	// Collect all the input IPs
	for pName, inPort := range p.InPorts() {
		inIPs[pName] = []*scipipe.FileIP{}
		for newIP := range inPort.Chan {
			inIPs[pName] = append(inIPs[pName], newIP)
		}
	}

	keys := []string{}
	for k := range inIPs {
		keys = append(keys, k)
	}

	outIPs := p.combine(inIPs, keys)

	// Send combinations of all IPs
	for pName, ips := range outIPs {
		for _, ip := range ips {
			p.Out(pName).Send(ip)
		}
	}
}

func (p *FileCombinator) combine(inIPs map[string][]*scipipe.FileIP, keys []string) map[string][]*scipipe.FileIP {
	if len(inIPs) <= 1 {
		return inIPs
	}

	headKey := keys[0]
	head := inIPs[headKey]

	tailKeys := keys[1:]
	tail := map[string][]*scipipe.FileIP{}
	for _, k := range tailKeys {
		tail[k] = inIPs[k]
	}
	tail = p.combine(tail, tailKeys) // Recursive call

	outIPs := map[string][]*scipipe.FileIP{}
	outIPs[headKey] = []*scipipe.FileIP{}
	for _, ip := range head {
		// Multiply each string in head with the length of the rows in the tail
		// (they are guaranteed to be of equal length)
		for i := 0; i < len(tail[tailKeys[0]]); i++ {
			outIPs[headKey] = append(outIPs[headKey], ip)
		}
		// Multiply the content of each row in the tail with the number of rows
		// in the tail
		for j := 0; j < len(tail); j++ {
			if len(outIPs) <= j {
				outIPs[tailKeys[j]] = []*scipipe.FileIP{}
			}
			outIPs[tailKeys[j]] = append(outIPs[tailKeys[j]], tail[tailKeys[j]]...)
		}
	}

	return outIPs
}
