package scipipe

import (
	"fmt"
	"os/exec"
	re "regexp"
	str "strings"
	// "time"
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
		typ := m[1]
		name := m[2]
		if typ == "o" {
			t.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
		} else if typ == "i" {
			// Set up a channel on the inports, even though this is
			// often replaced by another tasks output port channel.
			// It might be nice to have it init'ed with a channel
			// anyways, for use cases when we want to send FileTargets
			// on the inport manually.
			t.InPorts[name] = make(chan *FileTarget, BUFSIZE)
		}
	}
	return t
}

func (t *ShellTask) Run() {
	// Close output channels
	for _, ochan := range t.OutPorts {
		defer close(ochan)
	}
	for {
		breakLoop := false
		if len(t.InPorts) == 0 {
			breakLoop = true
		}

		// Set up inport / path mappings
		for iname, ichan := range t.InPorts {
			infile, open := <-ichan
			if !open {
				fmt.Println("Setting breakLoop to true")
				breakLoop = true
				continue
			}
			fmt.Println("Infile:", infile.GetPath())
			t.InPaths[iname] = infile.GetPath()
		}

		// Execute command
		t.executeCommands(t.Command)

		// Send output targets
		for oname, ochan := range t.OutPorts {
			fn := t.OutPathFuncs[oname]
			baseName := fn()
			nf := NewFileTarget(baseName)
			fmt.Println("Sending file:", nf.GetPath())
			ochan <- nf
		}
		if breakLoop {
			fmt.Println("Exiting main loop of task", t.Command)
			break
		}
	}
}

func (t *ShellTask) executeCommands(cmd string) {
	cmd = t.ReplacePortDefsInCmd(cmd)
	fmt.Println("ShellTask: Executing command: ", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

func (t *ShellTask) ReplacePortDefsInCmd(cmd string) string {
	r, err := re.Compile("{(o|i):([^{}:]+)}")
	Check(err)
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		whole := m[0]
		typ := m[1]
		name := m[2]
		newstr := "REPLACE_FAILED_FOR_PORT_" + name + "_CHECK_YOUR_CODE"
		if typ == "o" {
			newstr = t.OutPathFuncs[name]()
		} else if typ == "i" {
			newstr = t.InPaths[name]
		}
		cmd = str.Replace(cmd, whole, newstr, -1)
	}
	return cmd
}

func (t *ShellTask) GetInPath(inPort string) string {
	inPath := t.InPaths[inPort]
	return inPath
}
