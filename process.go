package scipipe

import (
	"errors"
	str "strings"
)

// ExecMode specifies which execution mode should be used for a Process and
// its corresponding Tasks
type ExecMode int

const (
	// ExecModeLocal indicates that commands on the local computer
	ExecModeLocal ExecMode = iota
	// ExecModeSLURM indicates that commands should be executed on a HPC cluster
	// via a SLURM resource manager
	ExecModeSLURM ExecMode = iota
)

// ================== Process ==================

// Process is the central component in SciPipe after Workflow. Processes are
// long-running "services" that schedules and executes Tasks based on the IPs
// and parameters received on its in-ports and parameter ports
type Process struct {
	name             string
	CommandPattern   string
	ExecMode         ExecMode
	Prepend          string
	Spawn            bool
	inPorts          map[string]*InPort
	outPorts         map[string]*OutPort
	OutPortsDoStream map[string]bool
	PathFormatters   map[string]func(*Task) string
	paramInPorts     map[string]*ParamInPort
	CustomExecute    func(*Task)
	workflow         *Workflow
	CoresPerTask     int
}

// NewProcess returns a new Process (without initializing its ports based on the
// command pattern. If this is what you need, use NewProc instead)
func NewProcess(workflow *Workflow, name string, command string) *Process {
	p := &Process{
		name:             name,
		CommandPattern:   command,
		inPorts:          make(map[string]*InPort),
		outPorts:         make(map[string]*OutPort),
		OutPortsDoStream: make(map[string]bool),
		PathFormatters:   make(map[string]func(*Task) string),
		paramInPorts:     make(map[string]*ParamInPort),
		Spawn:            true,
		workflow:         workflow,
		CoresPerTask:     1,
	}
	workflow.AddProc(p)
	return p
}

// ----------- Main API init methods ------------

// NewProc returns a new Process, and initializes its ports based on the
// command pattern.
func NewProc(workflow *Workflow, name string, cmd string) *Process {
	InitLogInfo()
	p := NewProcess(workflow, name, cmd)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

// ShellExpand expands the command pattern in cmd with the concrete values
// provided in inPaths, ouPaths and params
func ShellExpand(workflow *Workflow, name string, cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *Process {
	cmdExpr := expandCommandParamsAndPaths(cmd, params, inPaths, outPaths)
	p := NewProcess(workflow, name, cmdExpr)
	p.initPortsFromCmdPattern(cmdExpr, params)
	return p
}

// ------------------------------------------------
// Main API methods
// ------------------------------------------------

// Name returns the name of the process
func (p *Process) Name() string {
	return p.name
}

// ------------------------------------------------
// In-port stuff
// ------------------------------------------------

// In returns the in-port with name portName
func (p *Process) In(portName string) *InPort {
	if p.inPorts[portName] == nil {
		Error.Fatalf("No such in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.inPorts[portName]
}

// SetInPort adds the in-port port to the process, with name portName
func (p *Process) SetInPort(portName string, port *InPort) {
	if p.inPorts[portName] != nil {
		Error.Fatalf("Such an in-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	p.inPorts[portName] = port
}

// InPorts returns a map of all the in-ports of the process, keyed by their
// names
func (p *Process) InPorts() map[string]*InPort {
	return p.inPorts
}

// ------------------------------------------------
// Out-port stuff
// ------------------------------------------------

// Out returns the out-port with name portName
func (p *Process) Out(portName string) *OutPort {
	if p.outPorts[portName] == nil {
		Error.Fatalf("No such out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	return p.outPorts[portName]
}

// SetOutPort adds the out-port port to the process, with name portName
func (p *Process) SetOutPort(portName string, port *OutPort) {
	if p.outPorts[portName] != nil {
		Error.Fatalf("Such an out-port ('%s') already exists for process '%s'. Please check your workflow code!\n", portName, p.name)
	}
	p.outPorts[portName] = port
}

// OutPorts returns a map of all the out-ports of the process, keyed by their
// names
func (p *Process) OutPorts() map[string]*OutPort {
	return p.outPorts
}

// ------------------------------------------------
// Param-in-port stuff
// ------------------------------------------------

// ParamInPort returns the parameter port with name paramPortName
func (p *Process) ParamInPort(paramPortName string) *ParamInPort {
	if p.paramInPorts[paramPortName] == nil {
		Error.Fatalf("No such param-port ('%s') for process '%s'. Please check your workflow code!\n", paramPortName, p.name)
	}
	return p.paramInPorts[paramPortName]
}

// ParamInPorts returns all parameter ports of the process
func (p *Process) ParamInPorts() map[string]*ParamInPort {
	return p.paramInPorts
}

// SetParamInPort adds the parameter port paramPort with name paramPortName
func (p *Process) SetParamInPort(paramPortName string, paramPort *ParamInPort) {
	p.paramInPorts[paramPortName] = paramPort
}

// ------------------------------------------------
// Param-out-port stuff
// ------------------------------------------------

// ParamOutPorts returns an empty map of ParamOutPorts, to compy with the
// WorkflowProcess interface
func (p *Process) ParamOutPorts() map[string]*ParamOutPort {
	return map[string]*ParamOutPort{}
}

// ------------------------------------------------
// Path formatting stuff
// ------------------------------------------------

// SetPathStatic creates an (output) path formatter returning a static string file name
func (p *Process) SetPathStatic(outPortName string, path string) {
	p.PathFormatters[outPortName] = func(t *Task) string {
		return path
	}
}

// SetPathExtend creates an (output) path formatter that extends the path of
// an input IP
func (p *Process) SetPathExtend(inPortName string, outPortName string,
	extension string) {
	p.PathFormatters[outPortName] = func(t *Task) string {
		return t.InPath(inPortName) + extension
	}
}

// SetPathReplace creates an (output) path formatter that uses an input's path
// but replaces parts of it.
func (p *Process) SetPathReplace(inPortName string, outPortName string, old string, new string) {
	p.PathFormatters[outPortName] = func(t *Task) string {
		return str.Replace(t.InPath(inPortName), old, new, -1)
	}
}

// SetPathCustom takes a function which produces a file path based on data
// available in *Task, such as concrete file paths and parameter values,
func (p *Process) SetPathCustom(outPortName string, pathFmtFunc func(task *Task) (path string)) {
	p.PathFormatters[outPortName] = pathFmtFunc
}

// ------- Helper methods for initialization -------

// ExpandParams takes a command pattern and a map of parameter names mapped to
// parameter values, and returns the command as a string where any parameter
// placeholders (on the form `{p:paramname}` are replaced with the parameter
// value from the provided parameter values map.
func ExpandParams(cmd string, params map[string]string) string {
	return expandCommandParamsAndPaths(cmd, params, nil, nil)
}

func expandCommandParamsAndPaths(cmd string, params map[string]string, inPaths map[string]string, outPaths map[string]string) (cmdExpr string) {
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	if params != nil {
		Debug.Println("Params:", params)
	}
	if inPaths != nil {
		Debug.Println("inPaths:", inPaths)
	}
	if outPaths != nil {
		Debug.Println("outPaths:", outPaths)
	}
	cmdExpr = cmd
	Debug.Println("Got command: ", cmd)
	for _, m := range ms {
		placeHolderStr := m[0]
		typ := m[1]
		name := m[2]
		var filePath string
		if typ == "p" {
			if params != nil {
				if val, ok := params[name]; ok {
					Debug.Println("Found param:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = str.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		} else if typ == "i" {
			if inPaths != nil {
				if val, ok := inPaths[name]; ok {
					Debug.Println("Found inPath:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = str.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		} else if typ == "o" || typ == "os" {
			if outPaths != nil {
				if val, ok := outPaths[name]; ok {
					Debug.Println("Found outPath:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = str.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		}
	}
	if cmd != cmdExpr {
		Debug.Printf("Expanded command '%s' into '%s'\n", cmd, cmdExpr)
	}
	return
}

// Set up in- and out-ports based on the shell command pattern used to create the
// Process. Ports are set up in this way:
// `{i:PORTNAME}` specifies an in-port
// `{o:PORTNAME}` specifies an out-port
// `{os:PORTNAME}` specifies an out-port that streams via a FIFO file
// `{p:PORTNAME}` a "parameter-port", which means a port where parameters can be "streamed"
func (p *Process) initPortsFromCmdPattern(cmd string, params map[string]string) {

	// Find in/out port names and Params and set up in struct fields
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	Debug.Printf("Got the following matches for placeholders in the command, when initializing ports: %v\n", ms)

	for _, m := range ms {
		if len(m) < 3 {
			msg := "Too few matches"
			Check(errors.New(msg), msg)
		}

		typ := m[1]
		name := m[2]
		if typ == "o" || typ == "os" {
			p.outPorts[name] = NewOutPort(name)
			p.outPorts[name].Process = p
			if typ == "os" {
				p.OutPortsDoStream[name] = true
			}
		} else if typ == "i" {
			// Set up a channel on the inports, even though this is
			// often replaced by another processes output port channel.
			// It might be nice to have it init'ed with a channel
			// anyways, for use cases when we want to send IP
			// on the inport manually.
			p.inPorts[name] = NewInPort(name)
			p.inPorts[name].Process = p
		} else if typ == "p" {
			if params == nil || params[name] == "" {
				p.paramInPorts[name] = NewParamInPort()
				p.paramInPorts[name].Process = p
			}
		}
	}
}

// ------- Sanity checks -------

// IsConnected checks whether all the process' ports are connected
func (p *Process) IsConnected() (isConnected bool) {
	isConnected = true
	for portName, port := range p.inPorts {
		if !port.IsConnected() {
			Error.Printf("InPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	for portName, port := range p.outPorts {
		if !port.IsConnected() {
			Error.Printf("OutPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	for portName, port := range p.paramInPorts {
		if !port.IsConnected() {
			Error.Printf("ParamInPort %s of process %s is not connected - check your workflow code!\n", portName, p.name)
			isConnected = false
		}
	}
	return isConnected
}

// ============== Process Run Method ===============

// Run runs the process by instantiating and executing Tasks for all inputs
// and parameter values on its in-ports. in the case when there are no inputs
// or parameter values on the in-ports, it will run just once before it
// terminates. note that the actual execution of shell commands are done inside
// Task.Execute, not here.
func (p *Process) Run() {
	// Check that CoresPerTask is a sane number
	if p.CoresPerTask > cap(p.workflow.concurrentTasks) {
		Error.Fatalf("%s: CoresPerTask (%d) can't be greater than maxConcurrentTasks of workflow (%d)\n", p.Name(), p.CoresPerTask, cap(p.workflow.concurrentTasks))
	}

	defer p.closeOutPorts()

	tasks := []*Task{}
	Debug.Printf("Process %s: Starting to create and schedule tasks\n", p.name)
	for t := range p.createTasks() {

		// Collect created tasks, for the second round
		// where tasks are waited for to finish, before
		// sending their outputs.
		Debug.Printf("Process %s: Instantiated task [%s] ...", p.name, t.Command)
		tasks = append(tasks, t)

		anyPreviousFifosExists := t.anyFifosExist()

		if p.ExecMode == ExecModeLocal {
			if !anyPreviousFifosExists {
				Debug.Printf("Process %s: No FIFOs existed, so creating, for task [%s] ...", p.name, t.Command)
				t.createFifos()
			}

			// Sending FIFOs for the task
			for oname, oip := range t.OutTargets {
				if oip.doStream {
					p.Out(oname).Send(oip)
				}
			}
		}

		if anyPreviousFifosExists {
			Debug.Printf("Process %s: Previous FIFOs existed, so not executing task [%s] ...\n", p.name, t.Command)
			// Since t.Execute() is not run, that normally sends the Done signal, we
			// have to send it manually here:
			go func() {
				defer close(t.Done)
				t.Done <- 1
			}()
		} else {
			Debug.Printf("Process %s: Go-Executing task in separate go-routine: [%s] ...\n", p.name, t.Command)
			// Run the task
			go t.Execute()
			Debug.Printf("Process %s: Done go-executing task in go-routine: [%s] ...\n", p.name, t.Command)
		}
	}

	Debug.Printf("Process %s: Starting to loop over %d tasks to send out targets ...\n", p.name, len(tasks))
	for _, t := range tasks {
		Debug.Printf("Process %s: Waiting for Done from task: [%s]\n", p.name, t.Command)
		<-t.Done
		Debug.Printf("Process %s: Received Done from task: [%s]\n", p.name, t.Command)
		for oname, oip := range t.OutTargets {
			if !oip.doStream {
				Debug.Printf("Process %s: Sending target on outport %s, for task [%s] ...\n", p.name, oname, t.Command)
				p.Out(oname).Send(oip)
				Debug.Printf("Process %s: Done sending target on outport %s, for task [%s] ...\n", p.name, oname, t.Command)
			}
		}
	}
}

// -------- Helper methods for the Run method ---------

func (p *Process) receiveInputs() (inTargets map[string]*IP, inPortsOpen bool) {
	inPortsOpen = true
	inTargets = make(map[string]*IP)
	// Read input targets on in-ports and set up path mappings
	for inpName, inPort := range p.inPorts {
		Debug.Printf("Process %s: Receieving on inPort %s ...", p.name, inpName)
		inTarget, open := <-inPort.Chan
		if !open {
			inPortsOpen = false
			continue
		}
		Debug.Printf("Process %s: Got inTarget %s ...", p.name, inTarget.GetPath())
		inTargets[inpName] = inTarget
	}
	return
}

func (p *Process) receiveParams() (params map[string]string, paramPortsOpen bool) {
	paramPortsOpen = true
	params = make(map[string]string)
	// Read input targets on in-ports and set up path mappings
	for pname, pport := range p.paramInPorts {
		pval, open := <-pport.Chan
		if !open {
			paramPortsOpen = false
			continue
		}
		Debug.Println("Receiving param:", pname, "with value", pval)
		params[pname] = pval
	}
	return
}

func (p *Process) createTasks() (ch chan *Task) {
	ch = make(chan *Task)
	go func() {
		defer close(ch)
		for {
			inTargets, inPortsOpen := p.receiveInputs()
			Debug.Printf("Process.createTasks:%s Got inTargets: %v", p.name, inTargets)
			params, paramPortsOpen := p.receiveParams()
			Debug.Printf("Process.createTasks:%s Got params: %s", p.name, params)
			if !inPortsOpen && !paramPortsOpen {
				Debug.Printf("Process.createTasks:%s Breaking: Both inPorts and paramInPorts closed", p.name)
				break
			}
			if len(p.inPorts) == 0 && !paramPortsOpen {
				Debug.Printf("Process.createTasks:%s Breaking: No inports, and params closed", p.name)
				break
			}
			if len(p.paramInPorts) == 0 && !inPortsOpen {
				Debug.Printf("Process.createTasks:%s Breaking: No params, and inPorts closed", p.name)
				break
			}
			t := NewTask(p.workflow, p.name, p.CommandPattern, inTargets, p.PathFormatters, p.OutPortsDoStream, params, p.Prepend, p.ExecMode, p.CoresPerTask)
			if p.CustomExecute != nil {
				t.CustomExecute = p.CustomExecute
			}
			ch <- t
			if len(p.inPorts) == 0 && len(p.paramInPorts) == 0 {
				Debug.Printf("Process.createTasks:%s Breaking: No inports nor params", p.name)
				break
			}
		}
		Debug.Printf("Process.createTasks:%s Did break", p.name)
	}()
	return ch
}

func (p *Process) closeOutPorts() {
	for oname, oport := range p.outPorts {
		Debug.Printf("Process %s: Closing port(s) %s ...\n", p.name, oname)
		oport.Close()
	}
}
