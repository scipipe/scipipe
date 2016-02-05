package scipipe

import (
	"errors"
	str "strings"
)

// ================== ShellProcess ==================

type ShellProcess struct {
	Process
	InPorts          map[string]chan *FileTarget
	OutPorts         map[string]chan *FileTarget
	OutPortsDoStream map[string]bool
	PathFormatters   map[string]func(*ShellTask) string
	ParamPorts       map[string]chan string
	CustomExecute    func(*ShellTask)
	Prepend          string
	CommandPattern   string
	Spawn            bool
}

func NewShellProcess(command string) *ShellProcess {
	return &ShellProcess{
		CommandPattern:   command,
		InPorts:          make(map[string]chan *FileTarget),
		OutPorts:         make(map[string]chan *FileTarget),
		OutPortsDoStream: make(map[string]bool),
		PathFormatters:   make(map[string]func(*ShellTask) string),
		ParamPorts:       make(map[string]chan string),
		Spawn:            true,
	}
}

// ----------- Short-hand init methods ------------

func Sh(cmd string) *ShellProcess {
	return Shell(cmd)
}

func ShExp(cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *ShellProcess {
	return ShellExpand(cmd, inPaths, outPaths, params)
}

// ----------- Main API init methods ------------

func Shell(cmd string) *ShellProcess {
	if !LogExists {
		InitLogInfo()
	}
	p := NewShellProcess(cmd)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

func ShellExpand(cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *ShellProcess {
	cmdExpr := expandCommandParamsAndPaths(cmd, params, inPaths, outPaths)
	p := NewShellProcess(cmdExpr)
	p.initPortsFromCmdPattern(cmdExpr, params)
	return p
}

// ----------- Other API methods ------------

// Convenience method to create an (output) path formatter returning a static string
func (p *ShellProcess) SetPathFormatterString(outPort string, path string) {
	p.PathFormatters[outPort] = func(t *ShellTask) string {
		return path
	}
}

// Convenience method to create an (output) path formatter that extends the path of
// and input FileTarget
func (p *ShellProcess) SetPathFormatterExtend(outPort string, inPort string, extension string) {
	p.PathFormatters[outPort] = func(t *ShellTask) string {
		return t.InTargets[inPort].GetPath() + extension
	}
}

// Convenience method to create an (output) path formatter that uses an input's path
// but replaces parts of it.
func (p *ShellProcess) SetPathFormatterReplace(outPort string, inPort string, old string, new string) {
	p.PathFormatters[outPort] = func(t *ShellTask) string {
		return str.Replace(t.InTargets[inPort].GetPath(), old, new, -1)
	}
}

// ------- Helper methods for initialization -------

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
// ShellProcess. Ports are set up in this way:
// `{i:PORTNAME}` specifies an in-port
// `{o:PORTNAME}` specifies an out-port
// `{os:PORTNAME}` specifies an out-port that streams via a FIFO file
// `{p:PORTNAME}` a "parameter-port", which means a port where parameters can be "streamed"
func (p *ShellProcess) initPortsFromCmdPattern(cmd string, params map[string]string) {

	// Find in/out port names and Params and set up in struct fields
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)

	for _, m := range ms {
		if len(m) < 3 {
			Check(errors.New("Too few matches"))
		}

		typ := m[1]
		name := m[2]
		if typ == "o" || typ == "os" {
			p.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
			if typ == "os" {
				p.OutPortsDoStream[name] = true
			}
		} else if typ == "i" {
			// Set up a channel on the inports, even though this is
			// often replaced by another processes output port channel.
			// It might be nice to have it init'ed with a channel
			// anyways, for use cases when we want to send FileTargets
			// on the inport manually.
			p.InPorts[name] = make(chan *FileTarget, BUFSIZE)
		} else if typ == "p" {
			if params == nil || params[name] == "" {
				p.ParamPorts[name] = make(chan string, BUFSIZE)
			}
		}
	}
}

// ============== ShellProcess Run Method ===============

// The Run method of ShellProcess takes care of instantiating tasks for all
// sets of inputs and parameters that it receives, except when there is no
// input or parameter at all, when it will run once, and then terminate.
// Note that the actual execution of shell commands are done inside
// ShellTask, not ShellProcess.
func (p *ShellProcess) Run() {
	defer p.closeOutPorts()

	tasks := []*ShellTask{}
	Debug.Printf("[ShellProcess: %s] Starting to create and schedule tasks\n", p.CommandPattern)
	for t := range p.createTasks() {
		// Collect created tasks, for the second round
		// where tasks are waited for to finish, before
		// sending their outputs.
		Debug.Printf("[ShellProcess: %s] Instantiated task [%s] ...", p.CommandPattern, t.Command)
		tasks = append(tasks, t)

		anyPreviousFifosExists := t.anyFifosExist()
		if !anyPreviousFifosExists {
			Debug.Printf("[ShellProcess: %s] No FIFOs existed, so creating, for task [%s] ...", p.CommandPattern, t.Command)
			t.createFifos()
		}

		// Sending FIFOs for the task
		for oname, otgt := range t.OutTargets {
			if otgt.doStream {
				Debug.Printf("[ShellProcess: %s] Sending FIFO target on outport '%s' for task [%s] ...\n", p.CommandPattern, oname, t.Command)
				p.OutPorts[oname] <- otgt
			}
		}

		if !anyPreviousFifosExists {
			Debug.Printf("[ShellProcess: %s] Go-Executing task in separate go-routine: [%s] ...\n", p.CommandPattern, t.Command)
			// Run the task
			go t.Execute()
			Debug.Printf("[ShellProcess: %s] Done go-executing task in go-routine: [%s] ...\n", p.CommandPattern, t.Command)
		} else {
			// Since t.Execute() is not run, that normally sends the Done signal, we
			// have to send it manually here:
			go func() {
				defer close(t.Done)
				t.Done <- 1
			}()
		}
	}

	Debug.Printf("[ShellProcess: %s] Starting to loop over %d tasks to send out targets ...\n", p.CommandPattern, len(tasks))
	for _, t := range tasks {
		Debug.Printf("[ShellProcess: %s] Waiting for Done from task: [%s]\n", p.CommandPattern, t.Command)
		<-t.Done
		Debug.Printf("[ShellProcess: %s] Received Done from task: [%s]\n", p.CommandPattern, t.Command)
		for oname, otgt := range t.OutTargets {
			if !otgt.doStream {
				Debug.Printf("[ShellProcess: %s] Sending target on outport %s, for task [%s] ...\n", p.CommandPattern, oname, t.Command)
				p.OutPorts[oname] <- otgt
				Debug.Printf("[ShellProcess: %s] Done sending target on outport %s, for task [%s] ...\n", p.CommandPattern, oname, t.Command)
			}
		}
	}
}

// -------- Helper methods for the Run method ---------

func (p *ShellProcess) receiveInputs() (inTargets map[string]*FileTarget, inPortsOpen bool) {
	inPortsOpen = true
	inTargets = make(map[string]*FileTarget)
	// Read input targets on in-ports and set up path mappings
	for iname, ichan := range p.InPorts {
		Debug.Printf("[ShellProcess: %s] Receieving on inPort %s ...", p.CommandPattern, iname)
		inTarget, open := <-ichan
		if !open {
			inPortsOpen = false
			continue
		}
		Debug.Printf("[ShellProcess: %s] Got inTarget %s ...", p.CommandPattern, inTarget.GetPath())
		inTargets[iname] = inTarget
	}
	return
}

func (p *ShellProcess) receiveParams() (params map[string]string, paramPortsOpen bool) {
	paramPortsOpen = true
	params = make(map[string]string)
	// Read input targets on in-ports and set up path mappings
	for pname, pchan := range p.ParamPorts {
		pval, open := <-pchan
		if !open {
			paramPortsOpen = false
			continue
		}
		Debug.Println("Receiving param:", pname, "with value", pval)
		params[pname] = pval
	}
	return
}

func (p *ShellProcess) createTasks() (ch chan *ShellTask) {
	ch = make(chan *ShellTask)
	go func() {
		defer close(ch)
		for {
			inTargets, inPortsOpen := p.receiveInputs()
			Debug.Printf("[ShellProcess.createTasks: %s] Got inTargets: %s", p.CommandPattern, inTargets)
			params, paramPortsOpen := p.receiveParams()
			Debug.Printf("[ShellProcess.createTasks: %s] Got params: %s", p.CommandPattern, params)
			if !inPortsOpen && !paramPortsOpen {
				Debug.Printf("[ShellProcess.createTasks: %s] Breaking: Both inPorts and paramPorts closed", p.CommandPattern)
				break
			}
			if len(p.InPorts) == 0 && !paramPortsOpen {
				Debug.Printf("[ShellProcess.createTasks: %s] Breaking: No inports, and params closed", p.CommandPattern)
				break
			}
			if len(p.ParamPorts) == 0 && !inPortsOpen {
				Debug.Printf("[ShellProcess.createTasks: %s] Breaking: No params, and inPorts closed", p.CommandPattern)
				break
			}
			t := NewShellTask(p.CommandPattern, inTargets, p.PathFormatters, p.OutPortsDoStream, params, p.Prepend)
			if p.CustomExecute != nil {
				t.CustomExecute = p.CustomExecute
			}
			ch <- t
			if len(p.InPorts) == 0 && len(p.ParamPorts) == 0 {
				Debug.Printf("[ShellProcess.createTasks: %s] Breaking: No inports nor params", p.CommandPattern)
				break
			}
		}
	}()
	return ch
}

func (p *ShellProcess) closeOutPorts() {
	for oname, oport := range p.OutPorts {
		Debug.Printf("[ShellProcess: %s] Closing port %s ...\n", p.CommandPattern, oname)
		close(oport)
	}
}
