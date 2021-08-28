package components

import (
	"fmt"
	"sync"

	"github.com/scipipe/scipipe"
)

// ParamCombinator takes a set of input params, and returns the same
// number of param streams, where the params are multiplied so as to
// guarantee that all combinations of the params in the streams are created.
// Input ports and corresponding out-ports (with the same port names) are
// created on demand, by accessing them with the p.InParam(PORTNAME) method.
// The corresponding out-porta can then be accessed with the same port name
// with p.OutParam(PORTNAME)
type ParamCombinator struct {
	scipipe.BaseProcess
}

// NewParamCombinator returns a new initialized ParamCombinator process
func NewParamCombinator(wf *scipipe.Workflow, name string) *ParamCombinator {
	p := &ParamCombinator{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
	}
	wf.AddProc(p)
	return p
}

// InParam returns the in-port with name pName. If it does not exist, it will create
// that in-port, and a corresponding out-port with the same port name.
func (p *ParamCombinator) InParam(pName string) *scipipe.InParamPort {
	if _, ok := p.InParamPorts()[pName]; !ok {
		p.InitInParamPort(p, pName)
	}
	// Initialize a corresponding outport with the same name, for each inport
	if _, ok := p.OutParamPorts()[pName]; !ok {
		p.InitOutParamPort(p, pName)
	}
	return p.InParamPort(pName)
}

// OutParam returns the outport
func (p *ParamCombinator) OutParam(pName string) *scipipe.OutParamPort { return p.OutParamPort(pName) }

// Run runs the ParamCombinator process
func (p *ParamCombinator) Run() {
	defer p.CloseAllOutPorts()

	inParams := map[string][]string{}

	// Collect all input params
	for pName, inPort := range p.InParamPorts() {
		inParams[pName] = []string{}
		for newParam := range inPort.Chan {
			inParams[pName] = append(inParams[pName], newParam)
		}
	}

	keys := []string{}
	for k := range inParams {
		keys = append(keys, k)
	}

	outIPs := combine(inParams, keys)

	// Send combinations of all IPs
	wg := &sync.WaitGroup{}
	for pName, params := range outIPs {
		wg.Add(1)
		// Make unique copy of variables for this iteration, so they don't get
		// overwritten on the next loop iteration
		pName := pName
		ps := params
		go func() {
			for _, param := range ps {
				p.OutParam(pName).Send(param)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// combine is a recursive method that creates combinations of all the IPs in the input IP arrays, such that:
// [a b]
// [1 2 3]
// ... will be turned into:
// [a a a b b b]
// [1 2 3 1 2 3]
// as an example.
func combine(inParams map[string][]string, keys []string) map[string][]string {
	if len(inParams) <= 1 {
		return inParams
	}

	headKey := keys[0]
	head := inParams[headKey]

	tailKeys := keys[1:]
	tail := map[string][]string{}
	for _, k := range tailKeys {
		tail[k] = inParams[k]
	}
	tail = combine(tail, tailKeys) // Recursive call

	outParams := map[string][]string{}
	outParams[headKey] = []string{}
	for _, ip := range head {
		// Multiply each string in head with the length of the rows in the tail
		// (they are guaranteed to be of equal length)
		for i := 0; i < len(tail[tailKeys[0]]); i++ {
			outParams[headKey] = append(outParams[headKey], ip)
		}
		// Multiply the content of each row in the tail with the number of rows
		// in the tail
		for j := 0; j < len(tail); j++ {
			if len(outParams) <= j {
				outParams[tailKeys[j]] = []string{}
			}
			outParams[tailKeys[j]] = append(outParams[tailKeys[j]], tail[tailKeys[j]]...)
		}
	}

	return outParams
}

func (p *ParamCombinator) Failf(msg string, parts ...interface{}) {
	p.Fail(fmt.Sprintf(msg, parts...))
}

func (p *ParamCombinator) Fail(msg interface{}) {
	scipipe.Failf("[Process:%s] %s", p.Name(), msg)
}
