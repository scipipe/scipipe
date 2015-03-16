package scipipe

import (
	"fmt"
	"os/exec"
	re "regexp"
	str "strings"
)

type ShellTask struct {
	_OutOnly     bool
	Task         // Include stuff from "Parent Class"
	Command      string
	OutPorts     map[string]chan *FileTarget
	OutPathFuncs map[string]func() string
}

func NewShellTask(command string, outOnly bool) *ShellTask {
	t := new(ShellTask)
	t.Command = command
	t._OutOnly = outOnly
	if !t._OutOnly {
		t.InPorts = make(map[string]chan *FileTarget)
		t.InPaths = make(map[string]string)
	}
	t.OutPorts = make(map[string]chan *FileTarget)
	t.OutPathFuncs = make(map[string]func() string)
	return t
}

func Sh(cmd string) *ShellTask {
	outOnly := false

	r, err := re.Compile(".*{i:([^{}:]+)}.*")
	check(err)
	if !r.MatchString(cmd) {
		outOnly = true
	}

	t := NewShellTask(cmd, outOnly)

	if t._OutOnly {
		// Find in/out port names, and set up in port lists
		r, err := re.Compile("{o:([^{}:]+)}")
		check(err)
		ms := r.FindAllStringSubmatch(cmd, -1)
		for _, m := range ms {
			name := m[1]
			t.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
		}
	} else {
		// Find in/out port names, and set up in port lists
		r, err := re.Compile("{(o|i):([^{}:]+)}")
		check(err)
		ms := r.FindAllStringSubmatch(cmd, -1)
		for _, m := range ms {
			typ := m[1]
			name := m[2]
			if typ == "o" {
				t.OutPorts[name] = make(chan *FileTarget, BUFSIZE)
			} else if typ == "i" {
				// TODO: Is this really needed? SHouldn't inport chans be coming from previous tasks?
				t.InPorts[name] = make(chan *FileTarget, BUFSIZE)
			}
		}
	}

	return t
}

func (t *ShellTask) Init() {
	go func() {
		if t._OutOnly {

			t.executeCommands(t.Command)

			// Send output targets
			for oname, ochan := range t.OutPorts {
				fn := t.OutPathFuncs[oname]
				baseName := fn()
				nf := NewFileTarget(baseName)
				ochan <- nf
				close(ochan)
			}
		} else {
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
		}
	}()
}

func (t *ShellTask) executeCommands(cmd string) {
	cmd = t.ReplacePortDefsInCmd(cmd)
	fmt.Println("ShellTask Init(): Executing command: ", cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	check(err)
}

func (t *ShellTask) ReplacePortDefsInCmd(cmd string) string {
	r, err := re.Compile("{(o|i):([^{}:]+)}")
	check(err)
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
