package scipipe

import (
	"os"
	"strings"
)

// Process is the central component in SciPipe after Workflow. Processes are
// long-running "services" that schedules and executes Tasks based on the IPs
// and parameters received on its in-ports and parameter ports
type Process struct {
	BaseProcess
	CommandPattern   string
	OutPortsDoStream map[string]bool
	PathFormatters   map[string]func(*Task) string
	CustomExecute    func(*Task)
	CoresPerTask     int
	Prepend          string
	Spawn            bool
}

// ------------------------------------------------------------------------
// Factory method(s)
// ------------------------------------------------------------------------

// NewProc returns a new Process, and initializes its ports based on the
// command pattern.
func NewProc(workflow *Workflow, name string, cmd string) *Process {
	p := &Process{
		BaseProcess: NewBaseProcess(
			workflow,
			name,
		),
		CommandPattern:   cmd,
		OutPortsDoStream: make(map[string]bool),
		PathFormatters:   make(map[string]func(*Task) string),
		Spawn:            true,
		CoresPerTask:     1,
	}
	workflow.AddProc(p)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

// initPortsFromCmdPattern is a helper function for NewProc, that sets up in-
// and out-ports based on the shell command pattern used to create the Process.
// Ports are set up in this way:
// `{i:PORTNAME}` specifies an in-port
// `{o:PORTNAME}` specifies an out-port
// `{os:PORTNAME}` specifies an out-port that streams via a FIFO file
// `{p:PORTNAME}` a "parameter (in-)port", which means a port where parameters can be "streamed"
func (p *Process) initPortsFromCmdPattern(cmd string, params map[string]string) {
	// Find in/out port names and params and set up ports
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	if len(ms) == 0 {
		Fail("No placeholders found in command: " + cmd)
	}
	for _, m := range ms {
		portType := m[1]
		portName := m[2]
		if portType == "o" || portType == "os" {
			p.outPorts[portName] = NewOutPort(portName)
			p.outPorts[portName].process = p
			if portType == "os" {
				p.OutPortsDoStream[portName] = true
			}
		} else if portType == "i" {
			p.inPorts[portName] = NewInPort(portName)
			p.inPorts[portName].process = p
		} else if portType == "p" {
			if params == nil || params[portName] == "" {
				p.paramInPorts[portName] = NewParamInPort(portName)
				p.paramInPorts[portName].process = p
			}
		}
	}
}

// ------------------------------------------------------------------------
// Main API methods: Port accessor methods
// ------------------------------------------------------------------------

// In is a short-form for InPort() (of BaseProcess), which works only on Process
// processes
func (p *Process) In(portName string) *InPort {
	if portName == "" && len(p.InPorts()) == 1 {
		for _, inPort := range p.InPorts() {
			return inPort // Return the (only) in-port available
		}
	}
	return p.InPort(portName)
}

// Out is a short-form for OutPort() (of BaseProcess), which works only on
// Process processes
func (p *Process) Out(portName string) *OutPort {
	if portName == "" && len(p.OutPorts()) == 1 {
		for _, outPort := range p.OutPorts() {
			return outPort // Return the (only) out-port available
		}
	}
	return p.OutPort(portName)
}

// ------------------------------------------------------------------------
// Main API methods: Configure path formatting
// ------------------------------------------------------------------------

// SetPathStatic creates an (output) path formatter returning a static string file name
func (p *Process) SetPathStatic(outPortName string, path string) {
	p.PathFormatters[outPortName] = func(t *Task) string {
		return path
	}
}

// SetPathExtend creates an (output) path formatter that extends the path of
// an input IP
func (p *Process) SetPathExtend(inPortName string, outPortName string, extension string) {
	p.PathFormatters[outPortName] = func(t *Task) string {
		return t.InPath(inPortName) + extension
	}
}

// SetPathReplace creates an (output) path formatter that uses an input's path
// but replaces parts of it.
func (p *Process) SetPathReplace(inPortName string, outPortName string, old string, new string) {
	p.PathFormatters[outPortName] = func(t *Task) string {
		return strings.Replace(t.InPath(inPortName), old, new, -1)
	}
}

// SetPathCustom takes a function which produces a file path based on data
// available in *Task, such as concrete file paths and parameter values,
func (p *Process) SetPathCustom(outPortName string, pathFmtFunc func(task *Task) (path string)) {
	p.PathFormatters[outPortName] = pathFmtFunc
}

// ------------------------------------------------------------------------
// Run method
// ------------------------------------------------------------------------

// Run runs the process by instantiating and executing Tasks for all inputs
// and parameter values on its in-ports. in the case when there are no inputs
// or parameter values on the in-ports, it will run just once before it
// terminates. note that the actual execution of shell commands are done inside
// Task.Execute, not here.
func (p *Process) Run() {
	defer p.CloseOutPorts()
	// Check that CoresPerTask is a sane number
	if p.CoresPerTask > cap(p.workflow.concurrentTasks) {
		Failf("%s: CoresPerTask (%d) can't be greater than maxConcurrentTasks of workflow (%d)\n", p.Name(), p.CoresPerTask, cap(p.workflow.concurrentTasks))
	}

	tasks := []*Task{}
	for t := range p.createTasks() {
		// Collect tasks so we can later wait for their done-signal before sending outputs
		tasks = append(tasks, t)

		// Sending FIFOs for the task
		for oname, oip := range t.OutIPs {
			if oip.doStream {
				if oip.FifoFileExists() {
					Fail("Fifo file exists, so exiting (clean up fifo files before restarting the workflow): ", oip.FifoPath())
				}
				oip.CreateFifo()
				p.Out(oname).Send(oip)
			}
		}

		// Execute task in separate go-routine
		go t.Execute()
	}

	// Wait for tasks to finish, in the order they were started (thus maintaining
	// order of IPs), and then sending output IPs
	for _, t := range tasks {
		<-t.Done
		for oname, oip := range t.OutIPs {
			if !oip.doStream { // Streaming (FIFO) outputs have been sent earlier
				p.Out(oname).Send(oip)
			}
			// Remove any FIFO file
			if oip.doStream && oip.FifoFileExists() {
				os.Remove(oip.FifoPath())
			}
		}
	}
}

// createTasks is a helper method for Run that creates tasks based on incoming
// IPs on in-ports, and feeds them to the Run method on the returned channel ch
func (p *Process) createTasks() (ch chan *Task) {
	ch = make(chan *Task)
	go func() {
		defer close(ch)

		inIPs := map[string]*FileIP{}
		params := map[string]string{}

		inPortsOpen := true
		paramPortsOpen := true
		for {
			// Only read on in-ports if we have any
			if len(p.inPorts) > 0 {
				inIPs, inPortsOpen = p.receiveOnInPorts()
				// If in-port is closed, that means we got the last params on last iteration, so break
				if !inPortsOpen {
					break
				}
			}
			// Only read on param in-ports if we have any
			if len(p.paramInPorts) > 0 {
				params, paramPortsOpen = p.receiveOnParamInPorts()
				// If param-port is closed, that means we got the last params on last iteration, so break
				if !paramPortsOpen {
					break
				}
			}

			// Create task and send on the channel we are about to return
			ch <- NewTask(p.workflow, p, p.Name(), p.CommandPattern, inIPs, p.PathFormatters, p.OutPortsDoStream, params, p.Prepend, p.CustomExecute, p.CoresPerTask)

			// If we have no in-ports nor param in-ports, we should break after the first iteration
			if len(p.inPorts) == 0 && len(p.paramInPorts) == 0 {
				break
			}
		}
	}()
	return ch
}
