package components

import (
	"fmt"
	"sync"

	"github.com/scipipe/scipipe"
)

// FileCombinator takes a set of input streams of FileIPs, and returns the same
// number of output streams, where the FileIPs are multiplied so as to
// guarantee that all combinations of the ips in the input streams are created.
// Input ports and corresponding out-ports (with the same port names) are
// created on demand, by accessing them with the p.In(PORTNAME) method.
// The corresponding out-porta can then be accessed with the same port name
// with p.Out(PORTNAME)
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

// In returns the in-port with name pName. If it does not exist, it will create
// that in-port, and a corresponding out-port with the same port name.
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

// Out returns the outport
func (p *FileCombinator) Out(pName string) *scipipe.OutPort { return p.OutPort(pName) }

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
	wg := &sync.WaitGroup{}
	for pName, ips := range outIPs {
		wg.Add(1)
		// Make unique copy of variables for this iteration, so they don't get
		// overwritten on the next loop iteration
		pName := pName
		ips := ips
		go func() {
			for _, ip := range ips {
				p.Out(pName).Send(ip)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// combine is a recursive method that creates combinations of all the IPs in the input IP arrays, such that:
// [a.txt b.txt]
// [1,txt 2.txt 3.txt]
// ... will be turned into:
// [a.txt a.txt a.txt b.txt b.txt b.txt]
// [1.txt 2.txt 3.txt 1.txt 2.txt 3.txt]
// as an example.
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

func (p *FileCombinator) Failf(msg string, parts ...interface{}) {
	p.Fail(fmt.Sprintf(msg, parts...))
}

func (p *FileCombinator) Fail(msg interface{}) {
	scipipe.Failf("[Process:%s] %s", p.Name(), msg)
}
