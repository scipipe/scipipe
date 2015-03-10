package scipipe

import (
	"fmt"
	"os/exec"
	re "regexp"
	str "strings"
)

// ****** ShellTask ******

type ShellTask struct {
	Task    // Include stuff from "Parent Class"
	Command string
}

func Sh(cmd string) *ShellTask {
	t := NewShellTask()
	t.Command = cmd

	// Find in/out port names, and set up in port lists
	r, err := re.Compile("{(o|i):([^{}:]+)}")
	check(err, "hej1")
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		typ := m[1]
		name := m[2]
		if typ == "o" {
			t.OutPorts[name] = make(chan *FileTarget)
		} else if typ == "i" {
			// TODO: Is this really needed? SHouldn't inport chans be coming from previous tasks?
			t.InPorts[name] = make(chan *FileTarget)
		}
	}

	return t
}

func (t *ShellTask) Init() {
	go func() {
		for {
			doClose := false
			// Set up inport / path mappings
			for iname, ichan := range t.InPorts {
				infile, open := <-ichan
				if !open {
					doClose = true
				} else {
					t.InPaths[iname] = infile.GetPath()
				}
			}
			if doClose {
				break
			}

			t.executeCommands(t.Command)

			// Send output targets
			for oname, ochan := range t.OutPorts {
				fn := t.OutPathFuncs[oname]
				baseName := fn()
				nf := NewFileTarget(baseName)
				ochan <- nf
				if doClose {
					close(ochan)
				}
			}
		}
	}()
}

func (t *ShellTask) executeCommands(cmd string) {
	cmd = t.ReplacePortDefsInCmd(cmd)
	fmt.Println("ShellTask Init(): Executing command: ", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	check(err, "hej2")
}

func (t *ShellTask) ReplacePortDefsInCmd(cmd string) string {
	r, err := re.Compile("{(o|i):([^{}:]+)}")
	check(err, "hej1")
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

func NewShellTask() *ShellTask {
	t := new(ShellTask)
	t.InPorts = make(map[string]chan *FileTarget)
	t.InPaths = make(map[string]string)
	t.OutPorts = make(map[string]chan *FileTarget)
	t.OutPathFuncs = make(map[string]func() string)
	return t
}

// ****** ShellTaskOutputOnly ******

type ShellTaskOutputOnly struct {
	ShellTask
}

func NewShellTaskOutPutOnly() *ShellTaskOutputOnly {
	t := new(ShellTaskOutputOnly)
	t.OutPorts = make(map[string]chan *FileTarget)
	t.OutPathFuncs = make(map[string]func() string)
	return t
}

func ShOut(cmd string) *ShellTaskOutputOnly {
	t := NewShellTaskOutPutOnly()
	t.Command = cmd

	// Find in/out port names, and set up in port lists
	r, err := re.Compile("{o:([^{}:]+)}")
	check(err, "hej1")
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		name := m[1]
		t.OutPorts[name] = make(chan *FileTarget)
	}

	return t
}

func (t *ShellTaskOutputOnly) Init() {
	go func() {
		t.executeCommands(t.Command)

		// Send output targets
		for oname, ochan := range t.OutPorts {
			fn := t.OutPathFuncs[oname]
			baseName := fn()
			nf := NewFileTarget(baseName)
			ochan <- nf
			close(ochan)
		}
	}()
}
