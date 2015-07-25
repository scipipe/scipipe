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

type ShellTask struct {
	task
	_OutOnly     bool
	InPorts      map[string]chan *FileTarget
	InPaths      map[string]string
	OutPorts     map[string]chan *FileTarget
	OutPathFuncs map[string]func() string
	ParamPorts   map[string]chan string
	Params       map[string]string
	Prepend      string
	Command      string
	Spawn        bool
}

func NewShellTask(command string) *ShellTask {
	return &ShellTask{
		Command:      command,
		InPorts:      make(map[string]chan *FileTarget),
		InPaths:      make(map[string]string),
		OutPorts:     make(map[string]chan *FileTarget),
		OutPathFuncs: make(map[string]func() string),
		ParamPorts:   make(map[string]chan string),
		Params:       make(map[string]string),
		Spawn:        true,
	}
}

func Sh(cmd string) *ShellTask {
	return Shell(cmd)
}

func Shell(cmd string) *ShellTask {
	if !LogExists {
		InitLogAudit()
	}
	t := NewShellTask(cmd)
	t.initPortsFromCmdPattern(cmd, nil)
	return t
}

func ShExp(cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *ShellTask {
	return ShellExpand(cmd, inPaths, outPaths, params)
}

func ShellExpand(cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *ShellTask {
	cmdExp := expandCommandParamsAndPaths(cmd, params, inPaths, outPaths)
	t := NewShellTask(cmdExp)
	t.initPortsFromCmdPattern(cmdExp, params)
	return t
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

func (t *ShellTask) initPortsFromCmdPattern(cmd string, params map[string]string) {
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
			t.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
		} else if typ == "p" {
			if params == nil {
				t.ParamPorts[name] = make(chan string, BUFSIZE)
			} else {
				t.Params[name] = params[name]
			}
		}

		// else if typ == "i" {
		// Set up a channel on the inports, even though this is
		// often replaced by another tasks output port channel.
		// It might be nice to have it init'ed with a channel
		// anyways, for use cases when we want to send FileTargets
		// on the inport manually.
		// t.InPorts[name] = make(chan *FileTarget, BUFSIZE)
		// }
	}
}

func (t *ShellTask) Run() {
	Debug.Println("Entering task:", t.Command)
	defer t.closeOutChans()

	wg := new(sync.WaitGroup)
	mx := new(sync.Mutex)
	sendWaitQueue := []map[string]chan int{}
	// Main loop
	for {
		inPortsClosed := t.receiveInputs()
		paramPortsClosed := t.receiveParams()

		if len(t.InPorts) == 0 && paramPortsClosed {
			Debug.Println("Closing loop: No inports, and param ports closed")
			break
		}
		if len(t.ParamPorts) == 0 && inPortsClosed {
			Debug.Println("Closing loop: No param ports, and inports closed")
			break
		}
		if inPortsClosed && paramPortsClosed {
			Debug.Println("Closing loop: Both inports and param ports closed")
			break
		}

		// This is important that it is created anew here, for thread-safety
		outTargets := t.createOutTargets()
		// Format
		cmd := t.formatCommand(t.Command, outTargets)

		if t.anyFileExists(outTargets) {
			Warn.Printf("Skipping task, one or more outputs already exist: '%s'\n", cmd)
		} else {
			Audit.Printf("Starting task: '%s'\n", cmd)
			if t.Spawn {
				wg.Add(1)
				beforeSendCh := make(chan int)
				afterSendCh := make(chan int)
				sendWaitQueue = append(sendWaitQueue, map[string](chan int){"before": beforeSendCh, "after": afterSendCh})
				go func() {
					t.executeCommand(cmd)
					t.atomizeTargets(outTargets, mx)
					beforeSendCh <- 1
					t.sendOutputs(outTargets, mx)
					afterSendCh <- 1
					close(beforeSendCh)
					close(afterSendCh)
					wg.Done()
				}()
			} else {
				t.executeCommand(cmd)
				t.atomizeTargets(outTargets, mx)
				t.sendOutputs(outTargets, mx)
			}
			Audit.Printf("Finished task: '%s'\n", cmd)
		}

		// If there are no inports, we know we should exit the loop
		// directly after executing the command, and sending the outputs
		if len(t.InPorts) == 0 && len(t.ParamPorts) == 0 {
			Debug.Printf("Closing after send: No inports or param ports (task '%s')", cmd)
			break
		}
	}
	Debug.Printf("Starting to wait for ordered sends (task '%s')\n", t.Command)
	for i, sendChs := range sendWaitQueue {
		Debug.Printf("sendWaitQueue %d: Waiting to start sending ...\n", i)
		<-sendChs["before"]
		Debug.Printf("sendWaitQueue %d: Now starting to send ...\n", i)
		<-sendChs["after"]
		Debug.Printf("sendWaitQueue %d: Now has sent!\n", i)
	}
	Debug.Printf("Starting to wait (task '%s')\n", t.Command)
	wg.Wait()
	Debug.Printf("Finished waiting (task '%s')\n", t.Command)
	Debug.Println("Exiting task:", t.Command)
}

func (t *ShellTask) closeOutChans() {
	// Close output channels
	for _, ochan := range t.OutPorts {
		close(ochan)
	}
}

func (t *ShellTask) receiveInputs() bool {
	inPortsClosed := false
	// Read input targets on in-ports and set up path mappings
	for iname, ichan := range t.InPorts {
		infile, open := <-ichan
		if !open {
			inPortsClosed = true
			continue
		}
		Debug.Println("Receiving file:", infile.GetPath())
		t.InPaths[iname] = infile.GetPath()
	}
	return inPortsClosed
}

func (t *ShellTask) receiveParams() bool {
	paramPortsClosed := false
	// Read input targets on in-ports and set up path mappings
	for pname, pchan := range t.ParamPorts {
		pval, open := <-pchan
		if !open {
			paramPortsClosed = true
			continue
		}
		Debug.Println("Receiving param:", pname, "with value", pval)
		t.Params[pname] = pval
	}
	return paramPortsClosed
}

func (t *ShellTask) sendOutputs(outTargets map[string]*FileTarget, mx *sync.Mutex) {
	// Send output targets on out ports
	mx.Lock()
	for oname, ochan := range t.OutPorts {
		Debug.Println("Sending file:", outTargets[oname].GetPath())
		ochan <- outTargets[oname]
	}
	mx.Unlock()
}

func (t *ShellTask) createOutPaths() (outPaths map[string]string) {
	outPaths = make(map[string]string)
	for oname, ofun := range t.OutPathFuncs {
		outPaths[oname] = ofun()
	}
	return outPaths
}

func (t *ShellTask) createOutTargets() (outTargets map[string]*FileTarget) {
	outTargets = make(map[string]*FileTarget)
	for oname, opath := range t.createOutPaths() {
		outTargets[oname] = NewFileTarget(opath)
	}
	return
}

func (t *ShellTask) anyFileExists(targets map[string]*FileTarget) (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range targets {
		opath := tgt.GetPath()
		otmpPath := tgt.GetTempPath()
		if _, err := os.Stat(opath); err == nil {
			anyFileExists = true
			Debug.Println("Output file exists already:", opath)
		}
		if _, err := os.Stat(otmpPath); err == nil {
			anyFileExists = true
			Warn.Println("Temporary Output file already exists:", otmpPath, ". Check your workflow for correctness!")
		}
	}
	return
}

func (t *ShellTask) executeCommand(cmd string) {
	Info.Println("Executing cmd:", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

func (t *ShellTask) atomizeTargets(targets map[string]*FileTarget, mx *sync.Mutex) {
	mx.Lock()
	for _, tgt := range targets {
		Debug.Printf("Atomizing file: %s -> %s", tgt.GetTempPath(), tgt.GetPath())
		tgt.Atomize()
	}
	mx.Unlock()
}

func (t *ShellTask) formatCommand(cmd string, outTargets map[string]*FileTarget) string {
	r := getPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "o" {
			if outTargets[name] == nil {
				msg := fmt.Sprint("Missing outpath for outport '", name, "' of shell task '", t.Command, "'")
				Check(errors.New(msg))
			} else {
				newstr = outTargets[name].GetTempPath() // Means important to Atomize afterwards!
			}
		} else if typ == "i" {
			if t.InPaths[name] == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "' of shell task '", t.Command, "'")
				Check(errors.New(msg))
			} else {
				newstr = t.InPaths[name]
			}
		} else if typ == "p" {
			if t.Params[name] == "" {
				msg := fmt.Sprint("Missing param value param '", name, "' of shell task '", t.Command, "'")
				Check(errors.New(msg))
			} else {
				newstr = t.Params[name]
			}
		}
		if newstr == "" {
			msg := fmt.Sprint("Replace failed for port ", name, " in task '", t.Command, "'")
			Check(errors.New(msg))
		}
		cmd = str.Replace(cmd, whole, newstr, -1)
	}
	// Add prepend string to the command
	if t.Prepend != "" {
		cmd = fmt.Sprintf("%s %s", t.Prepend, cmd)
	}
	return cmd
}

func (t *ShellTask) GetInPath(inPort string) string {
	var inPath string
	if t.InPaths[inPort] != "" {
		inPath = t.InPaths[inPort]
	} else {
		msg := fmt.Sprint("t.GetInPath(): Missing inpath for inport '", inPort, "' of shell task '", t.Command, "'")
		Check(errors.New(msg))
	}
	return inPath
}

func getPlaceHolderRegex() *re.Regexp {
	r, err := re.Compile("{(o|i|p):([^{}:]+)}")
	Check(err)
	return r
}
