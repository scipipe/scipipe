package scipipe

import (
	"errors"
	"strings"
)

// ================== Process ==================

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
	ExecMode         ExecMode
	Prepend          string
	Spawn            bool
}

// NewProc returns a new Process, and initializes its ports based on the
// command pattern.
func NewProc(workflow *Workflow, name string, cmd string) *Process {
	p := newProcess(workflow, name, cmd)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

// newProcess returns a new Process (without initializing its ports based on the
// command pattern. If this is what you need, use NewProc instead)
func newProcess(workflow *Workflow, name string, command string) *Process {
	p := &Process{
		BaseProcess: NewBaseProcess(
			workflow,
			name,
		),
		CommandPattern:   command,
		OutPortsDoStream: make(map[string]bool),
		PathFormatters:   make(map[string]func(*Task) string),
		Spawn:            true,
		CoresPerTask:     1,
	}
	workflow.AddProc(p)
	return p
}

// ShellExpand expands the command pattern in cmd with the concrete values
// provided in inPaths, ouPaths and params
func ShellExpand(wf *Workflow, name string, cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *Process {
	cmdExpr := expandCommandParamsAndPaths(cmd, params, inPaths, outPaths)
	p := newProcess(wf, name, cmdExpr)
	p.initPortsFromCmdPattern(cmdExpr, params)
	return p
}

// ------------------------------------------------
// Main Process API methods
// ------------------------------------------------

// In is a short-form for InPort() (of BaseProcess), which works only on Process
// processes
func (p *Process) In(portName string) *InPort {
	return p.InPort(portName)
}

// Out is a short-form for OutPort() (of BaseProcess), which works only on
// Process processes
func (p *Process) Out(portName string) *OutPort {
	return p.OutPort(portName)
}

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
		return strings.Replace(t.InPath(inPortName), old, new, -1)
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
					cmdExpr = strings.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		} else if typ == "i" {
			if inPaths != nil {
				if val, ok := inPaths[name]; ok {
					Debug.Println("Found inPath:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = strings.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		} else if typ == "o" || typ == "os" {
			if outPaths != nil {
				if val, ok := outPaths[name]; ok {
					Debug.Println("Found outPath:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = strings.Replace(cmdExpr, placeHolderStr, filePath, -1)
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
			CheckWithMsg(errors.New(msg), msg)
		}

		typ := m[1]
		name := m[2]
		if typ == "o" || typ == "os" {
			p.outPorts[name] = NewOutPort(name)
			p.outPorts[name].process = p
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
			p.inPorts[name].process = p
		} else if typ == "p" {
			if params == nil || params[name] == "" {
				p.paramInPorts[name] = NewParamInPort(name)
				p.paramInPorts[name].process = p
			}
		}
	}
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
		Failf("%s: CoresPerTask (%d) can't be greater than maxConcurrentTasks of workflow (%d)\n", p.Name(), p.CoresPerTask, cap(p.workflow.concurrentTasks))
	}

	defer p.CloseOutPorts()

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
			for oname, oip := range t.OutIPs {
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

	Debug.Printf("Process %s: Starting to loop over %d tasks to send out IPs ...\n", p.name, len(tasks))
	for _, t := range tasks {
		Debug.Printf("Process %s: Waiting for Done from task: [%s]\n", p.name, t.Command)
		<-t.Done
		Debug.Printf("Process %s: Received Done from task: [%s]\n", p.name, t.Command)
		for oname, oip := range t.OutIPs {
			if !oip.doStream {
				Debug.Printf("Process %s: Sending IPs on outport %s, for task [%s] ...\n", p.name, oname, t.Command)
				p.Out(oname).Send(oip)
				Debug.Printf("Process %s: Done sending IPs on outport %s, for task [%s] ...\n", p.name, oname, t.Command)
			}
		}
	}
}

// -------- Helper methods for the Run method ---------

func (p *Process) createTasks() (ch chan *Task) {
	ch = make(chan *Task)
	go func() {
		defer close(ch)
		for {
			inIPs, inPortsOpen := p.receiveOnInPorts()
			Debug.Printf("Process.createTasks:%s Got inIPs: %v", p.name, inIPs)
			params, paramPortsOpen := p.receiveOnParamInPorts()
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
			t := NewTask(p.workflow, p.name, p.CommandPattern, inIPs, p.PathFormatters, p.OutPortsDoStream, params, p.Prepend, p.ExecMode, p.CoresPerTask)
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
