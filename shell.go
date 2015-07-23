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
	if !LogExists {
		InitLogError()
	}
	t := NewShellTask(cmd)
	t.initPortsFromCmdPattern(cmd, nil)
	return t
}

func ShParams(cmd string, params map[string]string) *ShellTask {
	t := NewShellTask(cmd)
	t.initPortsFromCmdPattern(cmd, params)
	return t
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
	Debug.Println("Entering task: ", t.Command)
	defer t.closeOutChans()

	wg := new(sync.WaitGroup)
	mx := new(sync.Mutex)
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
			Debug.Printf("No outputs exists, so starting task: '%s'\n", cmd)

			if t.Spawn {
				wg.Add(1)
				go func() {
					t.executeCommand(cmd)
					t.atomizeTargets(outTargets)
					t.sendOutputs(outTargets, mx)
					wg.Done()
				}()
			} else {
				t.executeCommand(cmd)
				t.atomizeTargets(outTargets)
				t.sendOutputs(outTargets, mx)
			}
		}

		// If there are no inports, we know we should exit the loop
		// directly after executing the command, and sending the outputs
		if len(t.InPorts) == 0 && len(t.ParamPorts) == 0 {
			Debug.Println("Closing after send: No inports or param ports")
			break
		}
	}
	Debug.Printf("Starting to wait (task '%s')\n", t.Command)
	wg.Wait()
	Debug.Printf("Finished waiting (task '%s')\n", t.Command)
	Debug.Println("Exiting task:  ", t.Command)
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
		Debug.Println("Sending file:  ", outTargets[oname].GetPath())
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
			Info.Println("Output file exists already:", opath)
		}
		if _, err := os.Stat(otmpPath); err == nil {
			anyFileExists = true
			Warn.Println("Temporary Output file already exists:", otmpPath, ". Check your workflow for correctness!")
		}
	}
	return
}

func (t *ShellTask) executeCommand(cmd string) {
	Info.Println("Executing cmd: ", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

func (t *ShellTask) atomizeTargets(targets map[string]*FileTarget) {
	for _, tgt := range targets {
		Debug.Printf("Atomizing file: %s -> %s", tgt.GetTempPath(), tgt.GetPath())
		tgt.Atomize()
	}
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
