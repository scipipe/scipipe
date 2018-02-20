package scipipe

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	str "strings"
	"time"
)

// Task represents a single static shell command, or go function, to be
// executed, and are scheduled and managed by a corresponding Process
type Task struct {
	Name          string
	Command       string
	ExecMode      ExecMode // TODO: Probably implement via different struct types (local/slurm/k8s, etc etc)
	CustomExecute func(*Task)
	InTargets     map[string]*FileIP
	OutTargets    map[string]*FileIP
	Params        map[string]string
	Done          chan int
	Image         string // TODO: Later probably only include in k8s-task
	DataFolder    string // TODO: Later probably only include in k8s-task
	workflow      *Workflow
	cores         int
}

// NewTask instantiates and initializes a new Task
func NewTask(workflow *Workflow, name string, cmdPat string, inTargets map[string]*FileIP, outPathFuncs map[string]func(*Task) string, outPortsDoStream map[string]bool, params map[string]string, prepend string, execMode ExecMode, cores int) *Task {
	t := &Task{
		Name:       name,
		InTargets:  inTargets,
		OutTargets: make(map[string]*FileIP),
		Params:     params,
		Command:    "",
		ExecMode:   execMode,
		Done:       make(chan int),
		workflow:   workflow,
		cores:      cores,
	}

	// Create out targets
	Debug.Printf("Task:%s: Creating outTargets now ... [%s]", name, cmdPat)
	outTargets := make(map[string]*FileIP)
	for oname, ofun := range outPathFuncs {
		opath := ofun(t)
		otgt := NewFileIP(opath)
		if outPortsDoStream[oname] {
			otgt.doStream = true
		}
		Debug.Printf("Task:%s: Creating outTarget with path %s ...\n", name, otgt.Path())
		outTargets[oname] = otgt
	}
	t.OutTargets = outTargets
	t.Command = formatCommand(cmdPat, inTargets, outTargets, params, prepend)
	Debug.Printf("Task:%s: Created formatted command: %s [%s]", name, t.Command, cmdPat)
	return t
}

// InPath returns the path name of an input file for the task
func (t *Task) InPath(portName string) string {
	if t.InTargets[portName] == nil {
		Error.Fatalf("No such portname (%s) in task (%s)\n", portName, t.Name)
	}
	return t.InTargets[portName].Path()
}

// Param returns the value of a param, for the task
func (t *Task) Param(portName string) string {
	if param, ok := t.Params[portName]; ok {
		return param
	}
	Error.Fatalf("No such param port '%s' for task '%s'\n", portName, t.Name)
	return "invalid"
}

// Execute executes the task (the shell command or go function configured for
// this task)
func (t *Task) Execute() {
	defer close(t.Done)

	if !t.anyOutputExists() && t.allFifosInOutTargetsExist() {
		Debug.Printf("Task:%-12s Executing task. [%s]\n", t.Name, t.Command)

		// Create directories for out-targets
		for _, oip := range t.OutTargets {
			oipDir := filepath.Dir(oip.Path())
			err := os.MkdirAll(oipDir, 0777)
			CheckWithMsg(err, "Could not create directory: "+oipDir)
		}

		t.workflow.IncConcurrentTasks(t.cores) // Will block if max concurrent tasks is reached
		startTime := time.Now()
		if t.CustomExecute != nil {
			Audit.Printf("Task:%-12s Executing custom execution function.\n", t.Name)
			t.CustomExecute(t)
		} else {
			switch t.ExecMode {
			case ExecModeLocal:
				t.executeCommand(t.Command)
			case ExecModeSLURM:
				Error.Printf("Task:%-12s SLURM Execution mode not implemented!", t.Name)
			}
		}
		execTime := time.Since(startTime)
		t.workflow.DecConcurrentTasks(t.cores)

		// Append audit info for the task to all its output targets

		auditInfo := NewAuditInfo()
		auditInfo.Command = t.Command
		auditInfo.Params = t.Params
		execTimeMS := execTime / time.Millisecond
		auditInfo.ExecTimeMS = execTimeMS
		// Set the audit infos from incoming IPs into the "Upstream" map
		for _, iip := range t.InTargets {
			iipPath := iip.Path()
			iipAuditInfo := iip.AuditInfo()
			auditInfo.Upstream[iipPath] = iipAuditInfo
		}
		// Add the current audit info to output ips and write them to file
		for _, oip := range t.OutTargets {
			oip.SetAuditInfo(auditInfo)
			for _, iip := range t.InTargets {
				oip.AddKeys(iip.Keys())
			}
			oip.WriteAuditLogToFile()
		}

		Debug.Printf("Task:%-12s Atomizing targets. [%s]\n", t.Name, t.Command)
		t.atomizeTargets()

	}
	Debug.Printf("Task:%s: Starting to send Done in t.Execute() ...) [%s]\n", t.Name, t.Command)
	t.Done <- 1
	Debug.Printf("Task:%s: Done sending Done, in t.Execute() [%s]\n", t.Name, t.Command)
}

// Check if any output file target, or temporary file targets, exist
func (t *Task) anyOutputExists() (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range t.OutTargets {
		opath := tgt.Path()
		otmpPath := tgt.TempPath()
		if !tgt.doStream {
			if _, err := os.Stat(opath); err == nil {
				Info.Printf("Task:%-12s Output file already exists, so skipping: %s\n", t.Name, opath)
				anyFileExists = true
			}
			if _, err := os.Stat(otmpPath); err == nil {
				Warning.Printf("Task:%-12s Temp   file already exists, so skipping: %s (Note: If resuming from a failed run, clean up .tmp files first).\n", t.Name, otmpPath)
				anyFileExists = true
			}
		}
	}
	return
}

// Check if any FIFO files for this tasks exist, for out-ports specified to support streaming
func (t *Task) anyFifosExist() (anyFifosExist bool) {
	anyFifosExist = false
	for _, tgt := range t.OutTargets {
		ofifoPath := tgt.FifoPath()
		if tgt.doStream {
			if _, err := os.Stat(ofifoPath); err == nil {
				Warning.Printf("Task:%-12s Output FIFO already exists, so skipping: %s (Note: If resuming from a failed run, clean up .fifo files first).\n", t.Name, ofifoPath)
				anyFifosExist = true
			}
		}
	}
	return
}

// Make sure that FIFOs that are supposed to exist, really exists
func (t *Task) allFifosInOutTargetsExist() bool {
	for _, tgt := range t.OutTargets {
		if tgt.doStream {
			if !tgt.FifoFileExists() {
				Warning.Printf("Task:%-12s FIFO Output file missing, for streaming output: %s. Check your workflow for correctness! [%s]\n", t.Name, t.Command, tgt.FifoPath())
				return false
			}
		}
	}
	return true
}

func (t *Task) executeCommand(cmd string) {
	Audit.Printf("Task:%-12s Executing command: %s\n", t.Name, cmd)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		Error.Printf("Command failed!\nCommand:\n%s\n\nOutput:\n%s\n\n", cmd, string(out))
		os.Exit(126)
	}
}

// Create FIFO files for all out-ports that are specified to support streaming
func (t *Task) createFifos() {
	Debug.Printf("Task:%s: Now creating fifos for task [%s]\n", t.Name, t.Command)
	for _, otgt := range t.OutTargets {
		if otgt.doStream {
			otgt.CreateFifo()
		}
	}
}

// Rename temporary output files to their proper file names
func (t *Task) atomizeTargets() {
	for _, tgt := range t.OutTargets {
		if !tgt.doStream {
			Debug.Printf("Atomizing file: %s -> %s", tgt.TempPath(), tgt.Path())
			tgt.Atomize()
			Debug.Printf("Done atomizing file: %s -> %s", tgt.TempPath(), tgt.Path())
		} else {
			Debug.Printf("Target is streaming, so not atomizing: %s", tgt.Path())
		}
	}
}

// Clean up any remaining FIFOs
// TODO: this is actually not really used anymore ...
func (t *Task) cleanUpFifos() {
	for _, tgt := range t.OutTargets {
		if tgt.doStream {
			Debug.Printf("Task:%s: Cleaning up FIFO for output target: %s [%s]\n", t.Name, tgt.FifoPath(), t.Command)
			tgt.RemoveFifo()
		} else {
			Debug.Printf("Task:%s: output target is not FIFO, so not removing any FIFO: %s [%s]\n", t.Name, tgt.Path(), t.Command)
		}
	}
}

var (
	trueVal  = true
	falseVal = false
)

func formatCommand(cmd string, inTargets map[string]*FileIP, outTargets map[string]*FileIP, params map[string]string, prepend string) string {
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	for _, m := range ms {
		reduceInputs := false

		placeHolderStr := m[0]
		typ := m[1]
		name := m[2]
		sep := " " // Default
		if len(m) > 3 {
			sep = m[5]
			reduceInputs = true
		}
		Debug.Printf("Found the following parts in the command: (type: '%s', name: '%s', sep: '%s', reduceInputs: %v). Command: %s\n", typ, name, sep, reduceInputs, cmd)
		var filePath string
		if typ == "o" || typ == "os" {
			// Out-ports
			if outTargets[name] == nil {
				msg := fmt.Sprint("Missing outpath for outport '", name, "' for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else {
				if typ == "o" {
					filePath = outTargets[name].TempPath() // Means important to Atomize afterwards!
				} else if typ == "os" {
					filePath = outTargets[name].FifoPath()
				}
			}
		} else if typ == "i" {
			// In-ports
			if inTargets[name] == nil {
				msg := fmt.Sprint("Missing intarget for inport '", name, "' for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else if inTargets[name].Path() == "" && reduceInputs {
				ips := []*FileIP{}
				for ip := range inTargets[name].SubStream.Chan {
					Debug.Println("Got ip: ", ip)
					ips = append(ips, ip)
				}
				Debug.Println("Got ips: ", ips)
				paths := []string{}
				Debug.Println("Got paths: ", paths)
				for _, ip := range ips {
					paths = append(paths, ip.Path())
				}
				Debug.Println("Got paths: ", paths)
				filePath = str.Join(paths, sep)
				Debug.Println("Got filePath: ", filePath)
			} else if inTargets[name].Path() == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "', and no substream, for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else {
				if inTargets[name].doStream {
					filePath = inTargets[name].FifoPath()
				} else {
					filePath = inTargets[name].Path()
				}
			}
			Debug.Printf("filePath determined to: %s, for command '%s'\n", filePath, cmd)
		} else if typ == "p" {
			if params[name] == "" {
				msg := fmt.Sprint("Missing param value param '", name, "' for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else {
				filePath = params[name]
			}
		}
		if filePath == "" {
			msg := fmt.Sprint("Replace failed for port ", name, " for command '", cmd, "'")
			CheckWithMsg(errors.New(msg), msg)
		}
		cmd = str.Replace(cmd, placeHolderStr, filePath, -1)
	}
	// Add prepend string to the command
	if prepend != "" {
		cmd = fmt.Sprintf("%s %s", prepend, cmd)
	}
	return cmd
}
