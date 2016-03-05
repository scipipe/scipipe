package scipipe

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	str "strings"
)

// ================== ShellTask ==================

type ShellTask struct {
	InTargets     map[string]*FileTarget
	OutTargets    map[string]*FileTarget
	Params        map[string]string
	Command       string
	CustomExecute func(*ShellTask)
	Done          chan int
}

func NewShellTask(cmdPat string, inTargets map[string]*FileTarget, outPathFuncs map[string]func(*ShellTask) string, outPortsDoStream map[string]bool, params map[string]string, prepend string) *ShellTask {
	t := &ShellTask{
		InTargets:  inTargets,
		OutTargets: make(map[string]*FileTarget),
		Params:     params,
		Command:    "",
		Done:       make(chan int),
	}
	// Create out targets
	Debug.Printf("[ShellTask: %s] Creating outTargets now ...", cmdPat)
	outTargets := make(map[string]*FileTarget)
	for oname, ofun := range outPathFuncs {
		opath := ofun(t)
		otgt := NewFileTarget(opath)
		if outPortsDoStream[oname] {
			otgt.doStream = true
		}
		Debug.Printf("[ShellTask: %s] Creating outTarget with path %s ...", cmdPat, opath)
		outTargets[oname] = otgt
	}
	t.OutTargets = outTargets
	t.Command = formatCommand(cmdPat, inTargets, outTargets, params, prepend)
	Debug.Printf("[ShellTask: %s] Created formatted command: %s", cmdPat, t.Command)
	return t
}

// --------------- ShellTask API methods ----------------

func (t *ShellTask) GetInPath(inPort string) string {
	return t.InTargets[inPort].GetPath()
}

func (t *ShellTask) Execute() {
	defer close(t.Done)
	if !t.anyOutputExists() && !t.fifosInOutTargetsMissing() {
		if t.CustomExecute != nil {
			Info.Printf("[Task: %s] Executing task.\n", t.Command)
			t.CustomExecute(t)
		} else {
			t.executeCommand(t.Command)
		}
		t.atomizeTargets()
	}
	Debug.Printf("[ShellTask: %s] Starting to send Done in t.Execute() ...)\n", t.Command)
	t.Done <- 1
	Debug.Printf("[ShellTask: %s] Done sending Done, in t.Execute()\n", t.Command)
}

// --------------- ShellTask Helper methods ----------------

// Check if any output file target, or temporary file targets, exist
func (t *ShellTask) anyOutputExists() (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range t.OutTargets {
		opath := tgt.GetPath()
		otmpPath := tgt.GetTempPath()
		if !tgt.doStream {
			if _, err := os.Stat(opath); err == nil {
				Warning.Printf("[ShellTask: %s] Output file already exists: %s, so skipping...\n", t.Command, opath)
				anyFileExists = true
			}
			if _, err := os.Stat(otmpPath); err == nil {
				Warning.Printf("[ShellTask: %s] Temporary Output file already exists: %s, so skipping...\n", t.Command, otmpPath)
				anyFileExists = true
			}
		}
	}
	return
}

// Check if any FIFO files for this tasks exist, for out-ports specified to support streaming
func (t *ShellTask) anyFifosExist() (anyFifosExist bool) {
	anyFifosExist = false
	for _, tgt := range t.OutTargets {
		ofifoPath := tgt.GetFifoPath()
		if tgt.doStream {
			if _, err := os.Stat(ofifoPath); err == nil {
				Warning.Printf("[ShellTask: %s] Output FIFO already exists: %s. Check your workflow for correctness!\n", t.Command, ofifoPath)
				anyFifosExist = true
			}
		}
	}
	return
}

// Make sure that FIFOs that are supposed to exist, really exists
func (t *ShellTask) fifosInOutTargetsMissing() (fifosInOutTargetsMissing bool) {
	fifosInOutTargetsMissing = false
	for _, tgt := range t.OutTargets {
		if tgt.doStream {
			ofifoPath := tgt.GetFifoPath()
			if _, err := os.Stat(ofifoPath); err != nil {
				Warning.Printf("[ShellTask: %s] FIFO Output file missing, for streaming output: %s. Check your workflow for correctness!\n", t.Command, ofifoPath)
				fifosInOutTargetsMissing = true
			}
		}
	}
	return
}

func (t *ShellTask) executeCommand(cmd string) {
	Info.Printf("[ShellTask: %s] Executing command: %s \n", t.Command, cmd)
	_, err := exec.Command("bash", "-c", cmd).Output()
	Check(err)
}

// Create FIFO files for all out-ports that are specified to support streaming
func (t *ShellTask) createFifos() {
	Debug.Printf("[ShellTask: %s] Now creating fifos for task\n", t.Command)
	for _, otgt := range t.OutTargets {
		if otgt.doStream {
			otgt.CreateFifo()
		}
	}
}

// Rename temporary output files to their proper file names
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

// Clean up any remaining FIFOs
// TODO: this is actually not really used anymore ...
func (t *ShellTask) cleanUpFifos() {
	for _, tgt := range t.OutTargets {
		if tgt.doStream {
			Debug.Printf("[ShellTask: %s] Cleaning up FIFO for output target: %s\n", t.Command, tgt.GetFifoPath())
			tgt.RemoveFifo()
		} else {
			Debug.Printf("[ShellTask: %s] output target is not FIFO, so not removing any FIFO: %s\n", t.Command, tgt.GetPath())
		}
	}
}

// ================== Helper functions==================

func formatCommand(cmd string, inTargets map[string]*FileTarget, outTargets map[string]*FileTarget, params map[string]string, prepend string) string {

	// Debug.Println("Formatting command with the following data:")
	// Debug.Println("prepend:", prepend)
	// Debug.Println("cmd:", cmd)
	// Debug.Println("inTargets:", inTargets)
	// Debug.Println("outTargets:", outTargets)
	// Debug.Println("params:", params)

	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		placeHolderStr := m[0]
		typ := m[1]
		name := m[2]
		var filePath string
		if typ == "o" || typ == "os" {
			// Out-ports
			if outTargets[name] == nil {
				msg := fmt.Sprint("Missing outpath for outport '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else {
				if typ == "o" {
					filePath = outTargets[name].GetTempPath() // Means important to Atomize afterwards!
				} else if typ == "os" {
					filePath = outTargets[name].GetFifoPath()
				}
			}
		} else if typ == "i" {
			// In-ports
			if inTargets[name] == nil {
				msg := fmt.Sprint("Missing intarget for inport '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else if inTargets[name].GetPath() == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else {
				if inTargets[name].doStream {
					filePath = inTargets[name].GetFifoPath()
				} else {
					filePath = inTargets[name].GetPath()
				}
			}
		} else if typ == "p" {
			if params[name] == "" {
				msg := fmt.Sprint("Missing param value param '", name, "' for command '", cmd, "'")
				Check(errors.New(msg))
			} else {
				filePath = params[name]
			}
		}
		if filePath == "" {
			msg := fmt.Sprint("Replace failed for port ", name, " for command '", cmd, "'")
			Check(errors.New(msg))
		}
		cmd = str.Replace(cmd, placeHolderStr, filePath, -1)
	}
	// Add prepend string to the command
	if prepend != "" {
		cmd = fmt.Sprintf("%s %s", prepend, cmd)
	}
	return cmd
}
