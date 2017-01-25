package scipipe

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	str "strings"
)

// ================== SciTask ==================

type SciTask struct {
	Name          string
	Command       string
	CustomExecute func(*SciTask)
	InTargets     map[string]*InformationPacket
	OutTargets    map[string]*InformationPacket
	Params        map[string]string
	Done          chan int
}

func NewSciTask(name string, cmdPat string, inTargets map[string]*InformationPacket, outPathFuncs map[string]func(*SciTask) string, outPortsDoStream map[string]bool, params map[string]string, prepend string) *SciTask {
	t := &SciTask{
		Name:       name,
		InTargets:  inTargets,
		OutTargets: make(map[string]*InformationPacket),
		Params:     params,
		Command:    "",
		Done:       make(chan int),
	}
	// Create out targets
	Debug.Printf("Task:%s: Creating outTargets now ... [%s]", name, cmdPat)
	outTargets := make(map[string]*InformationPacket)
	for oname, ofun := range outPathFuncs {
		opath := ofun(t)
		otgt := NewInformationPacket(opath)
		if outPortsDoStream[oname] {
			otgt.doStream = true
		}
		Debug.Printf("Task:%s: Creating outTarget with path %s ...\n", name, otgt.GetPath())
		outTargets[oname] = otgt
	}
	t.OutTargets = outTargets
	t.Command = formatCommand(cmdPat, inTargets, outTargets, params, prepend)
	Debug.Printf("Task:%s: Created formatted command: %s [%s]", name, t.Command, cmdPat)
	return t
}

// --------------- SciTask API methods ----------------

func (t *SciTask) GetInPath(inPort string) string {
	return t.InTargets[inPort].GetPath()
}

func (t *SciTask) Execute() {
	defer close(t.Done)
	if !t.anyOutputExists() && !t.fifosInOutTargetsMissing() {
		Debug.Printf("Task:%-12s Executing task. [%s]\n", t.Name, t.Command)
		if t.CustomExecute != nil {
			Audit.Printf("Task:%-12s Executing custom execution function.\n", t.Name)
			t.CustomExecute(t)
		} else {
			t.executeCommand(t.Command)
		}

		// Append audit info for the task to all its output targets
		auditInfo := NewAuditInfo()
		auditInfo.Command = t.Command
		auditInfo.Params = t.Params
		for _, oip := range t.OutTargets {
			oip.AddAuditInfo(auditInfo)
			oip.WriteAuditLogToFile()
		}

		Debug.Printf("Task:%-12s Atomizing targets. [%s]\n", t.Name, t.Command)
		t.atomizeTargets()
	}
	Debug.Printf("Task:%s: Starting to send Done in t.Execute() ...) [%s]\n", t.Name, t.Command)
	t.Done <- 1
	Debug.Printf("Task:%s: Done sending Done, in t.Execute() [%s]\n", t.Name, t.Command)
}

// --------------- SciTask Helper methods ----------------

// Check if any output file target, or temporary file targets, exist
func (t *SciTask) anyOutputExists() (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range t.OutTargets {
		opath := tgt.GetPath()
		otmpPath := tgt.GetTempPath()
		if !tgt.doStream {
			if _, err := os.Stat(opath); err == nil {
				Info.Printf("Task:%-12s Output file already exists, so skipping: %s\n", t.Name, opath)
				anyFileExists = true
			}
			if _, err := os.Stat(otmpPath); err == nil {
				Warning.Printf("Task:%-12s Temp   file already exists, so skipping: %s (Note: If resuming form a failed run, clean up .tmp files first).\n", t.Name, otmpPath)
				anyFileExists = true
			}
		}
	}
	return
}

// Check if any FIFO files for this tasks exist, for out-ports specified to support streaming
func (t *SciTask) anyFifosExist() (anyFifosExist bool) {
	anyFifosExist = false
	for _, tgt := range t.OutTargets {
		ofifoPath := tgt.GetFifoPath()
		if tgt.doStream {
			if _, err := os.Stat(ofifoPath); err == nil {
				Warning.Printf("Task:%-12s Output FIFO already exists, so skipping: %s (Note: If resuming form a failed run, clean up .fifo files first).\n", t.Name, ofifoPath)
				anyFifosExist = true
			}
		}
	}
	return
}

// Make sure that FIFOs that are supposed to exist, really exists
func (t *SciTask) fifosInOutTargetsMissing() (fifosInOutTargetsMissing bool) {
	fifosInOutTargetsMissing = false
	for _, tgt := range t.OutTargets {
		if tgt.doStream {
			ofifoPath := tgt.GetFifoPath()
			if _, err := os.Stat(ofifoPath); err != nil {
				Warning.Printf("Task:%-12s FIFO Output file missing, for streaming output: %s. Check your workflow for correctness! [%s]\n", t.Name, t.Command, ofifoPath)
				fifosInOutTargetsMissing = true
			}
		}
	}
	return
}

func (t *SciTask) executeCommand(cmd string) {
	Audit.Printf("Task:%-12s Executing command: %s\n", t.Name, cmd)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		Error.Println("Command failed, with output:\n", string(out))
		os.Exit(126)
	}
}

// Create FIFO files for all out-ports that are specified to support streaming
func (t *SciTask) createFifos() {
	Debug.Printf("Task:%s: Now creating fifos for task [%s]\n", t.Name, t.Command)
	for _, otgt := range t.OutTargets {
		if otgt.doStream {
			otgt.CreateFifo()
		}
	}
}

// Rename temporary output files to their proper file names
func (t *SciTask) atomizeTargets() {
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
func (t *SciTask) cleanUpFifos() {
	for _, tgt := range t.OutTargets {
		if tgt.doStream {
			Debug.Printf("Task:%s: Cleaning up FIFO for output target: %s [%s]\n", t.Name, tgt.GetFifoPath(), t.Command)
			tgt.RemoveFifo()
		} else {
			Debug.Printf("Task:%s: output target is not FIFO, so not removing any FIFO: %s [%s]\n", t.Name, tgt.GetPath(), t.Command)
		}
	}
}

// ================== Helper functions==================

func formatCommand(cmd string, inTargets map[string]*InformationPacket, outTargets map[string]*InformationPacket, params map[string]string, prepend string) string {

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
