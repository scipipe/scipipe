package scipipe

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	re "regexp"
	str "strings"
	"sync"
)

type ShellProcess struct {
	process
	InPorts          map[string]chan *FileTarget
	InPortsDoStream  map[string]bool
	InTargets        map[string]*FileTarget
	OutPorts         map[string]chan *FileTarget
	OutPortsDoStream map[string]bool
	OutPathFuncs     map[string]func() string
	ParamPorts       map[string]chan string
	Params           map[string]string
	Prepend          string
	Command          string
	Spawn            bool
}

func NewShellProcess(command string) *ShellProcess {
	return &ShellProcess{
		Command:          command,
		InPorts:          make(map[string]chan *FileTarget),
		InPortsDoStream:  make(map[string]bool),
		InTargets:        make(map[string]*FileTarget),
		OutPorts:         make(map[string]chan *FileTarget),
		OutPortsDoStream: make(map[string]bool),
		OutPathFuncs:     make(map[string]func() string),
		ParamPorts:       make(map[string]chan string),
		Params:           make(map[string]string),
		Spawn:            true,
	}
}

func Sh(cmd string) *ShellProcess {
	return Shell(cmd)
}

func Shell(cmd string) *ShellProcess {
	if !LogExists {
		InitLogAudit()
	}
	p := NewShellProcess(cmd)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

func ShExp(cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *ShellProcess {
	return ShellExpand(cmd, inPaths, outPaths, params)
}

func ShellExpand(cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *ShellProcess {
	cmdExp := expandCommandParamsAndPaths(cmd, params, inPaths, outPaths)
	p := NewShellProcess(cmdExp)
	p.initPortsFromCmdPattern(cmdExp, params)
	return p
}

func ShellPipable(cmd string, streamOutputs bool) {
	// TODO: Do stuff here
}

func expandCommandParamsAndPaths(cmd string, params map[string]string, inPaths map[string]string, outPaths map[string]string) (cmdExp string) {
	r := getPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	Debug.Println("Params:", params)
	Debug.Println("inPaths:", inPaths)
	Debug.Println("outPaths:", outPaths)
	cmdExp = cmd
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "p" {
			if params != nil && params[name] != "" {
				Debug.Println("Found param:", params[name])
				newstr = params[name]
			}
		} else if typ == "i" {
			if inPaths != nil && inPaths[name] != "" {
				Debug.Println("Found inPath:", inPaths[name])
				newstr = inPaths[name]
			}
		} else if typ == "o" {
			if outPaths != nil && outPaths[name] != "" {
				Debug.Println("Found outPath:", outPaths[name])
				newstr = outPaths[name]
			}
		}
		Debug.Println("Replacing:", whole, "->", newstr)
		cmdExp = str.Replace(cmdExp, whole, newstr, -1)
	}
	if cmd != cmdExp {
		Debug.Printf("Expanded command '%s' into '%s'\n", cmd, cmdExp)
	}
	return
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
		if typ == "o" {
			p.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
		} else if typ == "p" {
			if params == nil {
				p.ParamPorts[name] = make(chan string, BUFSIZE)
			} else {
				p.Params[name] = params[name]
			}
		}

		// else if typ == "i" {
		// Set up a channel on the inports, even though this is
		// often replaced by another processes output port channel.
		// It might be nice to have it init'ed with a channel
		// anyways, for use cases when we want to send FileTargets
		// on the inport manually.
		// p.InPorts[name] = make(chan *FileTarget, BUFSIZE)
		// }
	}
}

func (p *ShellProcess) Run() {
	Debug.Println("Entering process:", p.Command)
	defer p.closeOutChans()

	wg := new(sync.WaitGroup)
	mxCreateFifo := new(sync.Mutex)
	mxSendFifo := new(sync.Mutex)
	mxAtomize := new(sync.Mutex)
	mxSend := new(sync.Mutex)
	sendWaitQueue := []map[string]chan int{}
	// Main loop
	for {
		Debug.Printf("Looping again (%s)\n", p.Command)
		inPortsClosed := p.receiveInputs()
		paramPortsClosed := p.receiveParams()

		if len(p.InPorts) == 0 && paramPortsClosed {
			Debug.Println("Closing loop: No inports, and param ports closed")
			break
		}
		if len(p.ParamPorts) == 0 && inPortsClosed {
			Debug.Println("Closing loop: No param ports, and inports closed")
			break
		}
		if inPortsClosed && paramPortsClosed {
			Debug.Println("Closing loop: Both inports and param ports closed")
			break
		}
		Debug.Printf("Not closing (%s)\n", p.Command)

		// This is important that it is created anew here, for thread-safety
		outTargets := p.createOutTargets()
		// Format
		cmd := p.formatCommand(p.Command, outTargets)
		cmdForDisplay := str.Replace(cmd, ".tmp", "", -1)

		if p.anyFileExists(outTargets) {
			Warn.Printf("Skipping process, one or more outputs already exist: '%s'\n", cmd)
		} else {
			Audit.Printf("Starting shell task: %s\n", cmdForDisplay)
			if p.Spawn {
				Debug.Println("Task is spawned: ", cmdForDisplay)
				wg.Add(1)
				beforeFifoSendCh := make(chan int)
				afterFifoSendCh := make(chan int)
				beforeSendCh := make(chan int)
				afterSendCh := make(chan int)
				sendWaitQueue = append(sendWaitQueue, map[string](chan int){
					"beforefifo": beforeFifoSendCh,
					"afterfifo":  afterFifoSendCh,
					"before":     beforeSendCh,
					"after":      afterSendCh,
				})
				go func() {
					p.createFifos(outTargets, mxCreateFifo)
					beforeFifoSendCh <- 1
					p.sendFifoTargets(outTargets, mxSendFifo)
					afterFifoSendCh <- 1
					Debug.Println("Will execute command: ", cmd)
					p.executeCommand(cmd)
					p.atomizeTargets(outTargets, mxAtomize)
					beforeSendCh <- 1
					p.sendOutputs(outTargets, mxSend)
					afterSendCh <- 1
					close(beforeSendCh)
					close(afterSendCh)
					wg.Done()
				}()
			} else {
				p.createFifos(outTargets, mxCreateFifo)
				p.sendFifoTargets(outTargets, mxSendFifo)
				p.executeCommand(cmd)
				p.atomizeTargets(outTargets, mxAtomize)
				p.sendOutputs(outTargets, mxSend)
			}
			Audit.Printf("Finished shell task: %s\n", cmdForDisplay)
		}

		// If there are no inports, we know we should exit the loop
		// directly after executing the command, and sending the outputs
		if len(p.InPorts) == 0 && len(p.ParamPorts) == 0 {
			Debug.Printf("Closing loop after send: No inports or param ports (%s)", cmd)
			break
		}
	}
	Debug.Printf("Starting to wait for ordered sends (process '%s')\n", p.Command)
	for i, sendChs := range sendWaitQueue {
		Debug.Printf("sendWaitQueue %d: Waiting to start sending ...\n (%s)", i, p.Command)
		<-sendChs["beforefifo"]
		Debug.Printf("sendWaitQueue %d: Now starting to send fifos ...\n (%s)", i, p.Command)
		<-sendChs["afterfifo"]
		Debug.Printf("sendWaitQueue %d: Now has sent fifos!\n (%s)", i, p.Command)
		<-sendChs["before"]
		Debug.Printf("sendWaitQueue %d: Now starting to send ...\n (%s)", i, p.Command)
		<-sendChs["after"]
		Debug.Printf("sendWaitQueue %d: Now has sent!\n (%s)", i, p.Command)
	}
	Debug.Printf("Starting to wait (process '%s')\n", p.Command)
	wg.Wait()
	Debug.Printf("Finished waiting (process '%s')\n", p.Command)
	Debug.Println("Exiting process:", p.Command)
}

func (p *ShellProcess) closeOutChans() {
	// Close output channels
	for _, ochan := range p.OutPorts {
		close(ochan)
	}
}

func (p *ShellProcess) receiveInputs() bool {
	inPortsClosed := false
	// Read input targets on in-ports and set up path mappings
	for iname, ichan := range p.InPorts {
		inTarget, open := <-ichan
		if !open {
			inPortsClosed = true
			continue
		}
		Debug.Println("Receiving target:", inTarget.GetPath())
		p.InTargets[iname] = inTarget
	}
	return inPortsClosed
}

func (p *ShellProcess) receiveParams() bool {
	paramPortsClosed := false
	// Read input targets on in-ports and set up path mappings
	for pname, pchan := range p.ParamPorts {
		pval, open := <-pchan
		if !open {
			paramPortsClosed = true
			continue
		}
		Debug.Println("Receiving param:", pname, "with value", pval)
		p.Params[pname] = pval
	}
	return paramPortsClosed
}

func (p *ShellProcess) sendFifoTargets(outTargets map[string]*FileTarget, mx *sync.Mutex) {
	// Send output targets on out ports
	println("************************** HERE 1")
	mx.Lock()
	println("************************** HERE 2")
	for oname, ochan := range p.OutPorts {
		Debug.Println("Sending FIFO target:", outTargets[oname].GetPath())
		ochan <- outTargets[oname]
	}
	println("************************** HERE 3")
	mx.Unlock()
	println("************************** HERE 4")
}

func (p *ShellProcess) sendOutputs(outTargets map[string]*FileTarget, mx *sync.Mutex) {
	// Send output targets on out ports
	mx.Lock()
	for oname, ochan := range p.OutPorts {
		Debug.Println("Sending file:", outTargets[oname].GetPath())
		ochan <- outTargets[oname]
	}
	mx.Unlock()
}

func (p *ShellProcess) createOutPaths() (outPaths map[string]string) {
	outPaths = make(map[string]string)
	for oname, ofun := range p.OutPathFuncs {
		outPaths[oname] = ofun()
	}
	return outPaths
}

func (p *ShellProcess) createOutTargets() (outTargets map[string]*FileTarget) {
	outTargets = make(map[string]*FileTarget)
	for oname, opath := range p.createOutPaths() {
		outTargets[oname] = NewFileTarget(opath)
		// Set streaming mode on target (so can get picked up by downstream process)
		if p.OutPortsDoStream[oname] {
			outTargets[oname].doStream = true
		}
	}
	return
}

func (p *ShellProcess) anyFileExists(targets map[string]*FileTarget) (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range targets {
		opath := tgt.GetPath()
		otmpPath := tgt.GetTempPath()
		ofifoPath := tgt.GetFifoPath()
		if _, err := os.Stat(opath); err == nil {
			anyFileExists = true
			Debug.Println("Output file exists already:", opath)
		}
		if _, err := os.Stat(otmpPath); err == nil {
			anyFileExists = true
			Warn.Println("Temporary Output file already exists:", otmpPath, ". Check your workflow for correctness!")
		}
		if _, err := os.Stat(ofifoPath); err == nil {
			anyFileExists = true
			Warn.Println("FIFO Output file already exists:", otmpPath, ". Check your workflow for correctness!")
		}
	}
	return
}

func (p *ShellProcess) createFifos(outTargets map[string]*FileTarget, mx *sync.Mutex) {
	mx.Lock()
	for _, tgt := range outTargets {
		if tgt.doStream {
			Debug.Println("Creating FIFO:", tgt.GetFifoPath())
			tgt.CreateFifo()
		}
	}
}

func (p *ShellProcess) executeCommand(cmd string) {
	Info.Println("Executing cmd:", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

func (p *ShellProcess) atomizeTargets(targets map[string]*FileTarget, mx *sync.Mutex) {
	mx.Lock()
	for _, tgt := range targets {
		if !tgt.doStream {
			Debug.Printf("Atomizing file: %s -> %s", tgt.GetTempPath(), tgt.GetPath())
			tgt.Atomize()
		} else {
			Debug.Printf("Target is streaming, so not atomizing: %s", tgt.GetPath())
		}
	}
	mx.Unlock()
}

func (p *ShellProcess) formatCommand(cmd string, outTargets map[string]*FileTarget) string {
	r := getPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "o" {
			if outTargets[name] == nil {
				msg := fmt.Sprint("Missing outpath for outport '", name, "' of shell process '", p.Command, "'")
				Check(errors.New(msg))
			} else {
				if typ == "o" {
					newstr = outTargets[name].GetTempPath() // Means important to Atomize afterwards!
				} else if typ == "os" {
					newstr = outTargets[name].GetFifoPath()
				}
			}
		} else if typ == "i" || typ == "is" {
			if p.GetInPath(name) == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "' of shell process '", p.Command, "'")
				Check(errors.New(msg))
			} else {
				if typ == "i" {
					newstr = p.GetInPath(name)
				} else if typ == "is" {
					newstr = p.InTargets[name].GetFifoPath()
				}
			}
		} else if typ == "p" {
			if p.Params[name] == "" {
				msg := fmt.Sprint("Missing param value param '", name, "' of shell process '", p.Command, "'")
				Check(errors.New(msg))
			} else {
				newstr = p.Params[name]
			}
		}
		if newstr == "" {
			msg := fmt.Sprint("Replace failed for port ", name, " in process '", p.Command, "'")
			Check(errors.New(msg))
		}
		cmd = str.Replace(cmd, whole, newstr, -1)
	}
	// Add prepend string to the command
	if p.Prepend != "" {
		cmd = fmt.Sprintf("%s %s", p.Prepend, cmd)
	}
	return cmd
}

func getPlaceHolderRegex() *re.Regexp {
	r, err := re.Compile("{(o|i|p):([^{}:]+)}")
	Check(err)
	return r
}

func (p *ShellProcess) GetInPath(inPort string) string {
	var inPath string
	if p.InTargets[inPort] != nil {
		inPath = p.InTargets[inPort].GetPath()
	} else {
		msg := fmt.Sprint("p.GetInPath(): Missing inpath for inport '", inPort, "' of shell process '", p.Command, "'")
		Check(errors.New(msg))
	}
	return inPath
}

// Convenience method to create an (output) path formatter returning a static string
func (p *ShellProcess) OutPathGenString(outPort string, path string) {
	p.OutPathFuncs[outPort] = func() string {
		return path
	}
}

// Convenience method to create an (output) path formatter that extends the path of
// and input FileTarget
func (p *ShellProcess) OutPathGenExtend(outPort string, inPort string, extension string) {
	p.OutPathFuncs[outPort] = func() string {
		return p.GetInPath(inPort) + extension
	}
}

// Convenience method to create an (output) path formatter that uses an input's path
// but replaces parts of it.
func (p *ShellProcess) OutPathGenReplace(outPort string, inPort string, old string, new string) {
	p.OutPathFuncs[outPort] = func() string {
		return str.Replace(p.GetInPath(inPort), old, new, -1)
	}
}
