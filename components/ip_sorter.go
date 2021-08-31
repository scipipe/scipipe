package components

import (
	"sort"

	"github.com/scipipe/scipipe"
)

// IPSorter receives IPs and sorts them according to the string returned by
// sortingFunc. For example, if you want to sort by filepath, you can plug in a
// function that just returns ip.Path(), like so:
// func(ip *scipipe.FileIP) string { return ip.Path() }
// The process needs to wait for all IPs to arrive before sending any on the
// out-port, in order to be able to guarantee sorting.
type IPSorter struct {
	scipipe.BaseProcess
	sortingFunc func(*scipipe.FileIP) string
}

// NewIPSorter returns a new IPSorter component
func NewIPSorter(wf *scipipe.Workflow, name string, sortingFunc func(*scipipe.FileIP) string) *IPSorter {
	p := &IPSorter{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
	}
	p.InitInPort(p, "unsorted")
	p.InitOutPort(p, "sorted")
	wf.AddProc(p)
	return p
}

// In returns the in-port
func (p *IPSorter) In() *scipipe.InPort { return p.InPort("unsorted") }

// In returns the out-port
func (p *IPSorter) Out() *scipipe.OutPort { return p.OutPort("sorted") }

// Run runs the IPSorter component
func (p *IPSorter) Run() {
	defer p.CloseAllOutPorts()

	ips := map[string]*scipipe.FileIP{}
	sortStrings := []string{}

	// Read all IPs on in-port into data structures
	for inIP := range p.In().Chan {
		sortStr := inIP.Path()
		ips[sortStr] = inIP
		sortStrings = append(sortStrings, sortStr)
	}

	// Sort the strings representing IPs
	sort.Strings(sortStrings)

	// Send sorted IPs on out-port
	for _, sortStr := range sortStrings {
		p.Out().Send(ips[sortStr])
	}
}
