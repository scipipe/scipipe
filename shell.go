package scipipe

import (
	"fmt"
	"os/exec"
	re "regexp"
	str "strings"
	// "time"
	"errors"
)

type ShellTask struct {
	task
	_OutOnly     bool
	InPorts      map[string]chan *FileTarget
	InPaths      map[string]string
	OutPorts     map[string]chan *FileTarget
	OutPathFuncs map[string]func() string
	Command      string
}

func NewShellTask(command string) *ShellTask {
	return &ShellTask{
		Command:      command,
		InPorts:      make(map[string]chan *FileTarget),
		InPaths:      make(map[string]string),
		OutPorts:     make(map[string]chan *FileTarget),
		OutPathFuncs: make(map[string]func() string),
	}
}

func Sh(cmd string) *ShellTask {
	// Create task
	t := NewShellTask(cmd)

	// Find in/out port names, and set up in port lists
	r, err := re.Compile("{(o|i):([^{}:]+)}")
	Check(err)
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		if len(m) < 3 {
			Check(errors.New("Too few matches"))
		}
		typ := m[1]
		name := m[2]
		if typ == "o" {
			t.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
		} // else if typ == "i" {
		// Set up a channel on the inports, even though this is
		// often replaced by another tasks output port channel.
		// It might be nice to have it init'ed with a channel
		// anyways, for use cases when we want to send FileTargets
		// on the inport manually.
		// t.InPorts[name] = make(chan *FileTarget, BUFSIZE)
		// }
	}
	return t
}

func (t *ShellTask) Run() {
	fmt.Println("Entering task: ", t.Command)
	// Close output channels
	for _, ochan := range t.OutPorts {
		defer close(ochan)
	}

	// Main loop
	breakLoop := false
	for !breakLoop {
		// If there are no inports, we know we should exit the loop
		// directly after executing the command, and sending the outputs
		if len(t.InPorts) == 0 {
			breakLoop = true
		}

		// Read from inports
		inPortsOpen := t.receiveInputs()
		if !inPortsOpen {
			break
		}

		// Execute command
		t.formatAndExecute(t.Command)

		// Send
		t.sendOutputs()
	}
	fmt.Println("Exiting task:  ", t.Command)
}

func (t *ShellTask) receiveInputs() bool {
	inPortsOpen := true
	// Read input targets on in-ports and set up path mappings
	for iname, ichan := range t.InPorts {
		infile, open := <-ichan
		if !open {
			inPortsOpen = false
			continue
		}
		fmt.Println("Receiving file:", infile.GetPath())
		t.InPaths[iname] = infile.GetPath()
	}
	return inPortsOpen
}

func (t *ShellTask) sendOutputs() {
	// Send output targets on out ports
	for oname, ochan := range t.OutPorts {
		fun := t.OutPathFuncs[oname]
		baseName := fun()
		ft := NewFileTarget(baseName)
		fmt.Println("Sending file:  ", ft.GetPath())
		ochan <- ft
	}
}

func (t *ShellTask) formatAndExecute(cmd string) {
	cmd = t.ReplacePlaceholdersInCmd(cmd)
	fmt.Println("Executing cmd: ", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

func (t *ShellTask) ReplacePlaceholdersInCmd(cmd string) string {
	r, err := re.Compile("{(o|i):([^{}:]+)}")
	Check(err)
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		var newstr string
		if typ == "o" {
			if t.OutPathFuncs[name] != nil {
				newstr = t.OutPathFuncs[name]()
			} else {
				msg := fmt.Sprint("Missing outpath function for outport '", name, "' of shell task '", t.Command, "'")
				Check(errors.New(msg))
			}
		} else if typ == "i" {
			if t.InPaths[name] != "" {
				newstr = t.InPaths[name]
			} else {
				msg := fmt.Sprint("Missing inpath for inport '", name, "' of shell task '", t.Command, "'")
				Check(errors.New(msg))
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
		msg := fmt.Sprint("Missing inpath for inport '", inPort, "' of shell task '", t.Command, "'")
		Check(errors.New(msg))
	}
	return inPath
}
