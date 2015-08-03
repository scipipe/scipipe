package scipipe

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	re "regexp"
	str "strings"
)

type ShellProcess struct {
	process
	InPorts        map[string]chan *FileTarget
	OutPorts       map[string]chan *FileTarget
	OutPortsFifo   map[string]bool
	OutPathFuncs   map[string]func(*ShellTask) string
	ParamPorts     map[string]chan string
	Prepend        string
	CommandPattern string
	Spawn          bool
}

func NewShellProcess(command string) *ShellProcess {
	return &ShellProcess{
		CommandPattern: command,
		InPorts:        make(map[string]chan *FileTarget),
		OutPorts:       make(map[string]chan *FileTarget),
		OutPortsFifo:   make(map[string]bool),
		OutPathFuncs:   make(map[string]func(*ShellTask) string),
		ParamPorts:     make(map[string]chan string),
		Spawn:          true,
	}
}

func Shell(cmd string) *ShellProcess {
	if !LogExists {
		InitLogInfo()
	}
	p := NewShellProcess(cmd)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

func Sh(cmd string) *ShellProcess {
	return Shell(cmd)
}

func (p *ShellProcess) Run() {
	defer p.closeOutPorts()

	tasks := []*ShellTask{}
	Debug.Printf("[%s] Starting to loop over tasks\n", p.CommandPattern)
	for t := range p.createTasks() {
		Debug.Println("Now processing task", t.Command, "...")
		tasks = append(tasks, t)
		// Send fifos here
		go t.Run()
	}

	// Wait for finish, and send out targets in arrival order
	for _, t := range tasks {
		Debug.Printf("[%s] Waiting for Done from task: %s\n", p.CommandPattern, t.Command)
		<-t.Done
		Debug.Printf("[%s] Receiving Done from task: %s\n", p.CommandPattern, t.Command)
		for oname, otgt := range t.OutTargets {
			Debug.Printf("[%s] Sent on outport %s ...\n", p.CommandPattern, oname)
			p.OutPorts[oname] <- otgt
		}
	}
}

func (p *ShellProcess) initPortsFromCmdPattern(cmd string, params map[string]string) {
	// Find in/out port names and Params and set up in struct fields
	r := getPlaceHolderRegex()
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
				p.OutPortsFifo[name] = true
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

func (p *ShellProcess) createTasks() (ch chan *ShellTask) {
	ch = make(chan *ShellTask)
	go func() {
		defer close(ch)
		for {
			inTargets, inPortsOpen := p.receiveInputs()
			Debug.Printf("[%s] Got inTargets: %s", p.CommandPattern, inTargets)
			params, paramPortsOpen := p.receiveParams()
			Debug.Printf("[%s] Got params: %s", p.CommandPattern, params)
			if !inPortsOpen && !paramPortsOpen {
				Debug.Printf("[%s] Breaking: Both inPorts and paramPorts closed", p.CommandPattern)
				break
			}
			if len(p.InPorts) == 0 && !paramPortsOpen {
				Debug.Printf("[%s] Breaking: No inports, and params closed", p.CommandPattern)
				break
			}
			if len(p.ParamPorts) == 0 && !inPortsOpen {
				Debug.Printf("[%s] Breaking: No params, and inPorts closed", p.CommandPattern)
				break
			}
			t := NewShellTask(p.CommandPattern, inTargets, p.OutPathFuncs, params, p.Prepend)
			ch <- t
			if len(p.InPorts) == 0 && len(p.ParamPorts) == 0 {
				Debug.Printf("[%s] Breaking: No inports nor params", p.CommandPattern)
				break
			}
		}
	}()
	return ch
}

func (p *ShellProcess) receiveInputs() (inTargets map[string]*FileTarget, inPortsOpen bool) {
	inPortsOpen = true
	inTargets = make(map[string]*FileTarget)
	// Read input targets on in-ports and set up path mappings
	for iname, ichan := range p.InPorts {
		Debug.Printf("[%s] Receieving on inPort %s ...", p.CommandPattern, iname)
		inTarget, open := <-ichan
		if !open {
			inPortsOpen = false
			continue
		}
		Debug.Printf("[%s] Got inTarget %s ...", p.CommandPattern, inTarget.GetPath())
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

func (p *ShellProcess) closeOutPorts() {
	for oname, oport := range p.OutPorts {
		Debug.Printf("[%s] Closing port %s ...\n", p.CommandPattern, oname)
		close(oport)
	}
}

func getPlaceHolderRegex() *re.Regexp {
	r, err := re.Compile("{(o|os|i|is|p):([^{}:]+)}")
	Check(err)
	return r
}

// ------- ShellTask -------

type ShellTask struct {
	InTargets  map[string]*FileTarget
	OutTargets map[string]*FileTarget
	Params     map[string]string
	Command    string
	Done       chan int
}

func NewShellTask(cmdPat string, inTargets map[string]*FileTarget, outPathFuncs map[string]func(*ShellTask) string, params map[string]string, prepend string) *ShellTask {
	t := &ShellTask{
		InTargets:  inTargets,
		OutTargets: make(map[string]*FileTarget),
		Params:     params,
		Command:    "",
		Done:       make(chan int),
	}
	// Create out targets
	Debug.Printf("[%s] Creating outTargets now ...", cmdPat)
	outTargets := make(map[string]*FileTarget)
	for oname, ofun := range outPathFuncs {
		opath := ofun(t)
		Debug.Printf("[%s] Creating outTarget with path %s ...", cmdPat, opath)
		outTargets[oname] = NewFileTarget(opath)
	}
	t.OutTargets = outTargets
	t.Command = formatCommand(cmdPat, inTargets, outTargets, params, prepend)
	Debug.Printf("[%s] Created formatted command: %s", cmdPat, t.Command)
	return t
}

func (t *ShellTask) Run() {
	defer close(t.Done) // TODO: Is this needed?
	if !t.anyOutputExists() {
		t.executeCommand(t.Command)
		t.atomizeTargets()
	}
	Debug.Printf("[%s] Starting to send Done in t.Run() ...)\n", t.Command)
	t.Done <- 1
	Debug.Printf("[%s] Done sending Done, in t.Run()\n", t.Command)
}

func (t *ShellTask) executeCommand(cmd string) {
	Info.Printf("[%s] Executing command: %s \n", t.Command, cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

func (t *ShellTask) GetInPath(inPort string) string {
	return t.InTargets[inPort].GetPath()
}

func formatCommand(cmd string, inTargets map[string]*FileTarget, outTargets map[string]*FileTarget, params map[string]string, prepend string) string {

	// Debug.Println("Formatting command with the following data:")
	// Debug.Println("prepend:", prepend)
	// Debug.Println("cmd:", cmd)
	// Debug.Println("inTargets:", inTargets)
	// Debug.Println("outTargets:", outTargets)
	// Debug.Println("params:", params)

	r := getPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "o" || typ == "os" {
			if outTargets[name] == nil {
				msg := fmt.Sprint("Missing outpath for outport '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else {
				if typ == "o" {
					newstr = outTargets[name].GetTempPath() // Means important to Atomize afterwards!
				} else if typ == "os" {
					newstr = outTargets[name].GetFifoPath()
				}
			}
		} else if typ == "i" {
			if inTargets[name] == nil {
				msg := fmt.Sprint("Missing intarget for inport '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else if inTargets[name].GetPath() == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else {
				if typ == "i" {
					if inTargets[name].doStream {
						newstr = inTargets[name].GetFifoPath()
					} else {
						newstr = inTargets[name].GetPath()
					}
				}
			}
		} else if typ == "p" {
			if params[name] == "" {
				msg := fmt.Sprint("Missing param value param '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else {
				newstr = params[name]
			}
		}
		if newstr == "" {
			msg := fmt.Sprint("Replace failed for port ", name, " forcommand '", cmd, "'")
			Check(errors.New(msg))
		}
		cmd = str.Replace(cmd, whole, newstr, -1)
	}
	// Add prepend string to the command
	if prepend != "" {
		cmd = fmt.Sprintf("%s %s", prepend, cmd)
	}
	return cmd
}

func (t *ShellTask) atomizeTargets() {
	for _, tgt := range t.OutTargets {
		if !tgt.doStream {
			Debug.Printf("Atomizing file: %s -> %s", tgt.GetTempPath(), tgt.GetPath())
			tgt.Atomize()
			Debug.Printf("Done atomizing file: %s -> %s", tgt.GetTempPath(), tgt.GetPath())
		} else {
			Debug.Printf("Target is streaming, so not atomizing: %s", tgt.GetPath())
		}
	}
}

func (t *ShellTask) anyOutputExists() (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range t.OutTargets {
		opath := tgt.GetPath()
		otmpPath := tgt.GetTempPath()
		ofifoPath := tgt.GetFifoPath()
		if _, err := os.Stat(opath); err == nil {
			Warn.Printf("[%s] Output file already exists: %s. Check your workflow for correctness!\n", t.Command, opath)
			anyFileExists = true
		}
		if _, err := os.Stat(otmpPath); err == nil {
			Warn.Printf("[%s] Temporary Output file already exists: %s. Check your workflow for correctness!\n", t.Command, otmpPath)
			anyFileExists = true
		}
		if _, err := os.Stat(ofifoPath); err == nil {
			Warn.Printf("[%s] FIFO Output file already exists: %s. Check your workflow for correctness!\n", t.Command, ofifoPath)
			anyFileExists = true
		}
	}
	return
}
