package scipipe

import (
	"fmt"
	"os/exec"
	re "regexp"
	str "strings"
	// "time"
	"errors"
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
	}
}

func Sh(cmd string) *ShellTask {
	t := NewShellTask(cmd)
	t.initPortsFromCmdPattern(cmd)
	return t
}

func ShParams(cmd string, params map[string]string) *ShellTask {
	t := NewShellTask(cmd)
	t.initPortsFromCmdPattern(cmd)
	if params != nil {
		// Send eternal list of options
		go func() {
			for name, val := range params {
				t.ParamPorts[name] <- val
			}
		}()
	}
	return t
}

func (t *ShellTask) initPortsFromCmdPattern(cmd string) {
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
			t.ParamPorts[name] = make(chan string, BUFSIZE)
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
	// fmt.Println("Entering task: ", t.Command)
	defer t.closeOutChans()

	wg := new(sync.WaitGroup)
	mx := new(sync.Mutex)
	// Main loop
	for {
		inPortsClosed := t.receiveInputs()
		paramPortsClosed := t.receiveParams()

		if len(t.InPorts) == 0 && paramPortsClosed {
			// fmt.Println("Closing loop: No inports, and param ports closed")
			break
		}
		if len(t.ParamPorts) == 0 && inPortsClosed {
			// fmt.Println("Closing loop: No inports, and in ports closed")
			break
		}
		if inPortsClosed && paramPortsClosed {
			// fmt.Println("Closing loop: Both inports and param ports closed")
			break
		}

		// Really needed?
		outPaths := copyMapStrStr(t.createOutPaths())

		// Format
		cmd := t.formatCommand(t.Command)

		wg.Add(1)
		go func() {
			// Execute
			t.executeCommand(cmd)

			// Send
			t.sendOutputs(outPaths, mx)
			wg.Done()
		}()

		// If there are no inports, we know we should exit the loop
		// directly after executing the command, and sending the outputs
		if len(t.InPorts) == 0 && len(t.ParamPorts) == 0 {
			fmt.Println("Closing after send: No inports or param ports")
			break
		}
	}
	fmt.Printf("Starting to wait (task '%s')\n", t.Command)
	wg.Wait()
	fmt.Printf("Finished waiting (task '%s')\n", t.Command)
	// fmt.Println("Exiting task:  ", t.Command)
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
		// fmt.Println("Receiving file:", infile.GetPath())
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
		// fmt.Println("Receiving param:", pname, "with value", pval)
		t.Params[pname] = pval
	}
	return paramPortsClosed
}

func (t *ShellTask) sendOutputs(outPaths map[string]string, mx *sync.Mutex) {
	// Send output targets on out ports
	mx.Lock()
	for oname, ochan := range t.OutPorts {
		outName := outPaths[oname]
		ft := NewFileTarget(outName)
		// fmt.Println("Sending file:  ", ft.GetPath())
		ochan <- ft
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

func (t *ShellTask) executeCommand(cmd string) {
	fmt.Println("Executing cmd: ", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	// fmt.Println("Command output: ", string(cmdOut))
	Check(err)
}

func (t *ShellTask) formatCommand(cmd string) string {
	r := getPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "o" {
			if t.OutPathFuncs[name] == nil {
				msg := fmt.Sprint("Missing outpath function for outport '", name, "' of shell task '", t.Command, "'")
				Check(errors.New(msg))
			} else {
				newstr = t.OutPathFuncs[name]()
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
