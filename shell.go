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
	OutPorts         map[string]chan *FileTarget
	OutPortsDoStream map[string]bool
	OutPathFuncs     map[string]func(map[string]*FileTarget) string
	ParamPorts       map[string]chan string
	Prepend          string
	Command          string
	Spawn            bool
}

func NewShellProcess(command string) *ShellProcess {
	return &ShellProcess{
		Command:          command,
		InPorts:          make(map[string]chan *FileTarget),
		OutPorts:         make(map[string]chan *FileTarget),
		OutPortsDoStream: make(map[string]bool),
		OutPathFuncs:     make(map[string]func(map[string]*FileTarget) string),
		ParamPorts:       make(map[string]chan string),
		Spawn:            true,
	}
}

func Sh(cmd string) *ShellProcess {
	return Shell(cmd)
}

func Shell(cmd string) *ShellProcess {
	if !LogExists {
		InitLogInfo()
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

func expandCommandParamsAndPaths(cmd string, params map[string]string, inPaths map[string]string, outPaths map[string]string) (cmdExp string) {
	r := getPlaceHolderRegex()
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
	cmdExp = cmd
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "p" {
			if params != nil {
				if val, ok := params[name]; ok {
					Debug.Println("Found param:", val)
					newstr = val
					Debug.Println("Replacing:", whole, "->", newstr)
					cmdExp = str.Replace(cmdExp, whole, newstr, -1)
				}
			}
		} else if typ == "i" {
			if inPaths != nil {
				if val, ok := inPaths[name]; ok {
					Debug.Println("Found inPath:", val)
					newstr = val
					Debug.Println("Replacing:", whole, "->", newstr)
					cmdExp = str.Replace(cmdExp, whole, newstr, -1)
				}
			}
		} else if typ == "o" || typ == "os" {
			if outPaths != nil {
				if val, ok := outPaths[name]; ok {
					Debug.Println("Found outPath:", val)
					newstr = val
					Debug.Println("Replacing:", whole, "->", newstr)
					cmdExp = str.Replace(cmdExp, whole, newstr, -1)
				}
			}
		}
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
			if params == nil {
				p.ParamPorts[name] = make(chan string, BUFSIZE)
			}
		}
	}
}

func (p *ShellProcess) Run() {
	Debug.Println("Entering process:", p.Command)
	defer p.closeOutChans()

	wg := new(sync.WaitGroup)
	mxCreateFifo := new(sync.Mutex)
	mxSendFifo := new(sync.Mutex)
	mxAtomize := new(sync.Mutex)
	mxCleanUpFifos := new(sync.Mutex)
	mxSend := new(sync.Mutex)
	sendWaitQueue := []map[string]chan int{}
	// Main loop
	for {
		Debug.Printf("[%s] Looping again\n", p.Command)
		inTargets, inPortsClosed := p.receiveInputs()
		params, paramPortsClosed := p.receiveParams()

		// ----------------------------------------------------
		// Loop closing conditions
		// ----------------------------------------------------
		lenNonStreamingInports := 0
		for pname, _ := range p.InPorts {
			if !inTargets[pname].doStream {
				lenNonStreamingInports++
			}
		}
		if lenNonStreamingInports == 0 && paramPortsClosed {
			Debug.Printf("[%s] Closing loop after send: No non-streaming inports, and param ports closed\n", p.Command)
			break
		}
		if len(p.InPorts) == 0 && paramPortsClosed {
			Debug.Printf("[%s] Closing loop: No inports, and param ports closed\n", p.Command)
			break
		}
		if len(p.ParamPorts) == 0 && inPortsClosed {
			Debug.Printf("[%s] Closing loop: No param ports, and inports closed\n", p.Command)
			break
		}
		if inPortsClosed && paramPortsClosed {
			Debug.Printf("[%s] Closing loop: Both inports and param ports closed\n", p.Command)
			break
		}
		// ----------------------------------------------------
		// END: Loop closing conditions
		// ----------------------------------------------------

		// This is important that it is created anew here, for thread-safety
		outTargets := p.createOutTargets(inTargets)
		// Format
		cmd := p.formatCommand(p.Command, inTargets, outTargets, params)
		Debug.Printf("Formatted command: %s -> %s\n", p.Command, cmd)
		cmdForDisplay := getCmdForDisplay(cmd)

		if p.anyFileExists(outTargets) {
			Warn.Printf("[%s] Skipping process, one or more outputs already exist\n", cmd)
		} else {
			Info.Printf("[%s] Starting shell task\n", cmdForDisplay)
			if p.Spawn {
				Debug.Printf("[%s] Task is spawned\n", cmdForDisplay)
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
					Info.Printf("[%s] Starting execution of spawned shell task\n", cmdForDisplay)
					Debug.Printf("[%s] Creating fifos\n", cmdForDisplay)
					p.createFifos(outTargets, mxCreateFifo)
					Debug.Printf("[%s] Created fifos\n", cmdForDisplay)
					beforeFifoSendCh <- 1
					Debug.Printf("[%s] Sending fifos\n", cmdForDisplay)
					p.sendFifoTargets(outTargets, mxSendFifo)
					Debug.Printf("[%s] Sent fifos\n", cmdForDisplay)
					afterFifoSendCh <- 1
					Debug.Printf("[%s] Closing fifo channels\n", cmdForDisplay)
					close(beforeFifoSendCh)
					close(afterFifoSendCh)
					Debug.Printf("[%s] Closed fifo channels\n", cmdForDisplay)
					Debug.Printf("[%s] Will execute command\n", cmdForDisplay)
					p.executeCommand(cmd)
					p.atomizeTargets(outTargets, mxAtomize)
					p.cleanUpFifos(inTargets, mxCleanUpFifos)
					beforeSendCh <- 1
					p.sendOutputs(outTargets, mxSend)
					afterSendCh <- 1
					Debug.Printf("[%s] Closed output channels\n", cmdForDisplay)
					close(beforeSendCh)
					close(afterSendCh)
					Debug.Printf("[%s] Closed outputchannels\n", cmdForDisplay)
					Info.Printf("[%s] Finished execution of spawned shell task\n", cmdForDisplay)
					wg.Done()
				}()
			} else {
				Info.Printf("[%s] Starting execution of non-spawned shell task\n", cmdForDisplay)
				p.createFifos(outTargets, mxCreateFifo)
				p.sendFifoTargets(outTargets, mxSendFifo)
				p.executeCommand(cmd)
				p.atomizeTargets(outTargets, mxAtomize)
				p.cleanUpFifos(inTargets, mxCleanUpFifos)
				p.sendOutputs(outTargets, mxSend)
				Info.Printf("[%s] Finished execution of non-spawned shell task\n", cmdForDisplay)
			}
			Info.Printf("[%s] Finished spawning or execution of shell task\n", cmdForDisplay)
		}

		// If there are no inports, we know we should exit the loop
		// directly after executing the command, and sending the outputs
		if lenNonStreamingInports == 0 && len(p.ParamPorts) == 0 {
			Debug.Printf("[%s] Closing loop after send: No non-streaming inports, and no param ports\n", p.Command)
			break
		}
		if len(p.InPorts) == 0 && len(p.ParamPorts) == 0 {
			Debug.Printf("[%s] Closing loop after send: No inports or param ports\n", p.Command)
			break
		}
	}
	Debug.Printf("[%s] sendWaitQueue: Starting to wait for ordered sends\n", p.Command)
	for i, sendChs := range sendWaitQueue {
		Debug.Printf("[%s] sendWaitQueue %d: Waiting to start sending ...\n", i, p.Command)
		<-sendChs["beforefifo"]
		Debug.Printf("[%s] sendWaitQueue %d: Now starting to send fifos ...\n", i, p.Command)
		<-sendChs["afterfifo"]
		Debug.Printf("[%s] sendWaitQueue %d: Now has sent fifos!\n", i, p.Command)
		<-sendChs["before"]
		Debug.Printf("[%s] sendWaitQueue %d: Now starting to send ...\n", i, p.Command)
		<-sendChs["after"]
		Debug.Printf("[%s] sendWaitQueue %d: Now has sent!\n", i, p.Command)
	}
	Debug.Printf("[%s] Starting to wait for WaitGroup\n", p.Command)
	wg.Wait()
	Debug.Printf("[%s] Finished waiting\n", p.Command)
	Debug.Println("Exiting process:", p.Command)
}

func (p *ShellProcess) closeOutChans() {
	// Close output channels
	for _, ochan := range p.OutPorts {
		close(ochan)
	}
}

func (p *ShellProcess) receiveInputs() (inTargets map[string]*FileTarget, inPortsClosed bool) {
	inPortsClosed = false
	inTargets = make(map[string]*FileTarget)
	// Read input targets on in-ports and set up path mappings
	for iname, ichan := range p.InPorts {
		inTarget, open := <-ichan
		if !open {
			inPortsClosed = true
			continue
		}
		Debug.Println("Receiving target:", inTarget.GetPath())
		inTargets[iname] = inTarget
	}
	return
}

func (p *ShellProcess) receiveParams() (params map[string]string, paramPortsClosed bool) {
	paramPortsClosed = false
	params = make(map[string]string)
	// Read input targets on in-ports and set up path mappings
	for pname, pchan := range p.ParamPorts {
		pval, open := <-pchan
		if !open {
			paramPortsClosed = true
			continue
		}
		Debug.Println("Receiving param:", pname, "with value", pval)
		params[pname] = pval
	}
	return
}

func (p *ShellProcess) sendFifoTargets(outTargets map[string]*FileTarget, mx *sync.Mutex) {
	// Send output targets on out ports
	mx.Lock()
	for oname, ochan := range p.OutPorts {
		Debug.Println("Sending FIFO target:", outTargets[oname].GetPath())
		ochan <- outTargets[oname]
	}
	mx.Unlock()
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

func (p *ShellProcess) createOutPaths(inTargets map[string]*FileTarget) (outPaths map[string]string) {
	outPaths = make(map[string]string)
	for oname, ofun := range p.OutPathFuncs {
		outPaths[oname] = ofun(inTargets)
	}
	return outPaths
}

func (p *ShellProcess) createOutTargets(inTargets map[string]*FileTarget) (outTargets map[string]*FileTarget) {
	outTargets = make(map[string]*FileTarget)
	for oname, opath := range p.createOutPaths(inTargets) {
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
	mx.Unlock()
}

func (p *ShellProcess) executeCommand(cmd string) {
	cmdForDisplay := getCmdForDisplay(cmd)
	Info.Printf("[%s] Executing command: %s\n", cmdForDisplay, cmd)
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

func (p *ShellProcess) cleanUpFifos(inTargets map[string]*FileTarget, mx *sync.Mutex) {
	mx.Lock()
	for _, tgt := range inTargets {
		if tgt.doStream {
			Debug.Printf("[%s] Cleaning up FIFO for input target: %s\n", p.Command, tgt.GetFifoPath())
			tgt.RemoveFifo()
		} else {
			Debug.Printf("[%s] input target is not FIFO, so not removing any FIFO\n", tgt.GetPath())
		}
	}
	mx.Unlock()
}

func (p *ShellProcess) formatCommand(cmd string, inTargets map[string]*FileTarget, outTargets map[string]*FileTarget, params map[string]string) string {
	r := getPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "o" || typ == "os" {
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
		} else if typ == "i" {
			if inTargets[name].GetPath() == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "' of shell process '", p.Command, "'")
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
				msg := fmt.Sprint("Missing param value param '", name, "' of shell process '", p.Command, "'")
				Check(errors.New(msg))
			} else {
				newstr = params[name]
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
	r, err := re.Compile("{(o|os|i|is|p):([^{}:]+)}")
	Check(err)
	return r
}

func (p *ShellProcess) GetInPath(inPort string, inTargets map[string]*FileTarget) string {
	var inPath string
	if inTargets[inPort] != nil {
		inPath = inTargets[inPort].GetPath()
	} else {
		msg := fmt.Sprint("p.GetInPath(): Missing inpath for inport '", inPort, "' of shell process '", p.Command, "'")
		Check(errors.New(msg))
	}
	return inPath
}

// Convenience method to create an (output) path formatter returning a static string
func (p *ShellProcess) OutPathGenString(outPort string, path string) {
	p.OutPathFuncs[outPort] = func(inTargets map[string]*FileTarget) string {
		return path
	}
}

// Convenience method to create an (output) path formatter that extends the path of
// and input FileTarget
func (p *ShellProcess) OutPathGenExtend(outPort string, inPort string, extension string) {
	p.OutPathFuncs[outPort] = func(inTargets map[string]*FileTarget) string {
		return inTargets[inPort].GetPath() + extension
	}
}

// Convenience method to create an (output) path formatter that uses an input's path
// but replaces parts of it.
func (p *ShellProcess) OutPathGenReplace(outPort string, inPort string, old string, new string) {
	p.OutPathFuncs[outPort] = func(inTargets map[string]*FileTarget) string {
		return str.Replace(inTargets[inPort].GetPath(), old, new, -1)
	}
}

func getCmdForDisplay(cmd string) string {
	return str.Replace(str.Replace(cmd, ".tmp", "", -1), ".fifo", "", -1)
}
