package scipipe

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Process is the central component in SciPipe after Workflow. Processes are
// long-running "services" that schedules and executes Tasks based on the IPs
// and parameters received on its in-ports and parameter ports
type Process struct {
	BaseProcess
	CommandPattern string
	PathFuncs      map[string]func(*Task) string
	CustomExecute  func(*Task)
	CoresPerTask   int
	Prepend        string
	Spawn          bool
	PortInfo       map[string]*PortInfo
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
		CommandPattern: cmd,
		PathFuncs:      make(map[string]func(*Task) string),
		Spawn:          true,
		CoresPerTask:   1,
		PortInfo:       map[string]*PortInfo{},
	}
	workflow.AddProc(p)
	p.initPortsFromCmdPattern(cmd, nil)
	p.initDefaultPathFuncs()
	return p
}

// PortInfo is a container for various information about process ports
type PortInfo struct {
	portType  string
	extension string
	doStream  bool
	join      bool
	joinSep   string
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

	for _, m := range ms {
		portType := m[1]
		portRest := m[2]
		splitParts := strings.Split(portRest, "|")
		portName := splitParts[0]

		p.PortInfo[portName] = &PortInfo{portType: portType}

		for _, part := range splitParts[1:] {
			// If the |-separated part starts with a dot, treat it as a
			// configuration for file extenion to use
			fileExtPtn := regexp.MustCompile("\\.([a-z0-9\\.\\-\\_]+)")
			if fileExtPtn.MatchString(part) {
				m := fileExtPtn.FindStringSubmatch(part)
				p.PortInfo[portName].extension = m[1]
			}
			// If the |-separated part starts with "join:"
			// then treat the character following that as the character to use
			// when joining multiple files received on a sub-stream
			joinPtn := regexp.MustCompile("join:([^{}|]+)")
			if joinPtn.MatchString(part) {
				m := joinPtn.FindStringSubmatch(part)
				p.PortInfo[portName].join = true
				p.PortInfo[portName].joinSep = m[1]
			}
		}
	}

	for portName, pInfo := range p.PortInfo {
		if pInfo.portType == "o" || pInfo.portType == "os" {
			p.InitOutPort(p, portName)
			if pInfo.portType == "os" {
				p.PortInfo[portName].doStream = true
			}
		} else if pInfo.portType == "i" {
			p.InitInPort(p, portName)
		} else if pInfo.portType == "p" {
			if params == nil {
				p.InitInParamPort(p, portName)
			} else if _, ok := params[portName]; !ok {
				p.InitInParamPort(p, portName)
			}
		}
	}
}

// initDefaultPathFuncs does exactly what it name says: Initializes default
// path formatters for processes, that is used if no explicit path is set, using
// the proc.SetPath[...] methods
func (p *Process) initDefaultPathFuncs() {
	for outName := range p.OutPorts() {
		outName := outName
		p.PathFuncs[outName] = func(t *Task) string {
			pathPcs := []string{}
			for _, ipName := range sortedFileIPMapKeys(t.InIPs) {
				pathPcs = append(pathPcs, filepath.Base(t.InIP(ipName).Path()))
			}
			procName := sanitizePathFragment(t.Process.Name())
			pathPcs = append(pathPcs, procName)
			for _, paramName := range sortedStringMapKeys(t.Params) {
				pathPcs = append(pathPcs, paramName+"_"+t.Param(paramName))
			}
			for _, tagName := range sortedStringMapKeys(t.Tags) {
				pathPcs = append(pathPcs, tagName+"_"+t.Tag(tagName))
			}
			pathPcs = append(pathPcs, outName)
			fileExt := p.PortInfo[outName].extension
			if fileExt != "" {
				pathPcs = append(pathPcs, fileExt)
			}
			return strings.Join(pathPcs, ".")
		}
	}
}

func sortedFileIPMapKeys(kv map[string]*FileIP) []string {
	keys := []string{}
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringMapKeys(kv map[string]string) []string {
	keys := []string{}
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedFileIPSliceMapKeys(kv map[string][]*FileIP) []string {
	keys := []string{}
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

// InParam is a short-form for InParamPort() (of BaseProcess), which works only on Process
// processes
func (p *Process) InParam(portName string) *InParamPort {
	if _, ok := p.inParamPorts[portName]; !ok {
		p.InitInParamPort(p, portName)
	}
	return p.InParamPort(portName)
}

// OutParam is a short-form for OutParamPort() (of BaseProcess), which works only on
// Process processes
func (p *Process) OutParam(portName string) *OutParamPort {
	return p.OutParamPort(portName)
}

// ------------------------------------------------------------------------
// Main API methods: Configure path formatting
// ------------------------------------------------------------------------

// SetOut initializes a port (if it does not already exist), and takes a
// configuration for its outputs paths via a pattern similar to the command
// pattern used to create new processes, with placeholder tags. Available
// placeholder tags to use are:
// {i:inport_name}
// {p:param_name}
// {t:tag_name}
// An example might be: {i:foo}.replace_with_{p:replacement}.txt
// ... given that the process contains an in-port named 'foo', and a parameter
// named 'replacement'.
// If an out-port with the specified name does not exist, it will be created.
// This allows to create out-ports for filenames that are created without explicitly
// stating a filename on the commandline, such as when only submitting a prefix.
func (p *Process) SetOut(outPortName string, pathPattern string) {
	p.SetOutFunc(outPortName, func(t *Task) string {
		path := pathPattern // Avoiding reusing the same variable in multiple instances of this func

		r := getShellCommandPlaceHolderRegex()
		matches := r.FindAllStringSubmatch(path, -1)
		for _, match := range matches {
			var replacement string

			placeHolder := match[0]
			phType := match[1]
			restMatch := match[2]

			parts := strings.Split(restMatch, "|")
			portName := parts[0]
			restParts := parts[1:]

			switch phType {
			case "i":
				replacement = t.InPath(portName)
			case "o":
				if _, ok := t.Process.PathFuncs[portName]; !ok {
					Fail("No such pathfunc for out-port " + portName + " in task " + t.Name)
				}
				replacement = t.Process.PathFuncs[portName](t)
			case "p":
				replacement = t.Param(portName)
			case "t":
				replacement = t.Tag(portName)
			default:
				Fail("Replace failed for placeholder ", portName, " for path patterh '", path, "'")
			}

			if len(restParts) > 0 {
				replacement = applyPathModifiers(replacement, restParts)
			}

			// Replace placeholder with concrete value
			path = strings.Replace(path, placeHolder, replacement, -1)
		}
		return path
	})
}

// SetOutFunc takes a function which produces a file path based on data
// available in *Task, such as concrete file paths and parameter values,
func (p *Process) SetOutFunc(outPortName string, pathFmtFunc func(task *Task) (path string)) {
	if _, ok := p.outPorts[outPortName]; !ok {
		p.InitOutPort(p, outPortName)
	}
	p.PathFuncs[outPortName] = pathFmtFunc
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

	// Using a slice to store unprocessed tasks allows us to receive tasks as
	// they are produced and to maintain the correct order of IPs. This select
	// allows us to process completed tasks as they become available. Waiting
	// for all Tasks to be spawned before processing any can cause deadlock
	// under certain workflow architectures when there are more than getBufsize()
	// Tasks per process, see #81.
	startedTasks := taskQueue{}

	var nextTask *Task
	tasks := p.createTasks()
	for tasks != nil || len(startedTasks) > 0 {
		select {
		case t, ok := <-tasks:
			if !ok {
				tasks = nil
			} else {
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

				startedTasks = append(startedTasks, t)
			}
		case <-startedTasks.NextTaskDone():
			nextTask, startedTasks = startedTasks[0], startedTasks[1:]
			for oname, oip := range nextTask.OutIPs {
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
			// Tags need to be per Task, otherwise they are overwritten by future IPs
			tags := map[string]string{}
			// Only read on in-ports if we have any
			if len(p.inPorts) > 0 {
				inIPs, inPortsOpen = p.receiveOnInPorts()
				// If in-port is closed, that means we got the last params on last iteration, so break
				if !inPortsOpen {
					break
				}
			}
			// Only read on param in-ports if we have any
			if len(p.inParamPorts) > 0 {
				params, paramPortsOpen = p.receiveOnInParamPorts()
				// If param-port is closed, that means we got the last params on last iteration, so break
				if !paramPortsOpen {
					break
				}
			}

			for iname, ip := range inIPs {
				for k, v := range ip.Tags() {
					tags[iname+"."+k] = v
				}
			}

			// Create task and send on the channel we are about to return
			ch <- NewTask(p.workflow, p, p.Name(), p.CommandPattern, inIPs, p.PathFuncs, p.PortInfo, params, tags, p.Prepend, p.CustomExecute, p.CoresPerTask)

			// If we have no in-ports nor param in-ports, we should break after the first iteration
			if len(p.inPorts) == 0 && len(p.inParamPorts) == 0 {
				break
			}
		}
	}()
	return ch
}

type taskQueue []*Task

// NextTaskDone allows us to wait for the next task to be done if it's
// available. Otherwise, nil is returned since nil channels always block.
func (tq taskQueue) NextTaskDone() chan int {
	if len(tq) > 0 {
		return tq[0].Done
	}
	return nil
}
