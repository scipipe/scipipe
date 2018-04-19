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
	InIPs         map[string]*FileIP
	OutIPs        map[string]*FileIP
	Params        map[string]string
	Done          chan int
	Image         string // TODO: Later probably only include in k8s-task
	DataFolder    string // TODO: Later probably only include in k8s-task
	cores         int
	workflow      *Workflow
	process       *Process
}

// NewTask instantiates and initializes a new Task
func NewTask(workflow *Workflow, process *Process, name string, cmdPat string, inIPs map[string]*FileIP, outPathFuncs map[string]func(*Task) string, outPortsDoStream map[string]bool, params map[string]string, prepend string, execMode ExecMode, cores int) *Task {
	t := &Task{
		Name:     name,
		InIPs:    inIPs,
		OutIPs:   make(map[string]*FileIP),
		Params:   params,
		Command:  "",
		ExecMode: execMode,
		Done:     make(chan int),
		cores:    cores,
		workflow: workflow,
		process:  process,
	}

	// Create out IPs
	Debug.Printf("Task:%s: Creating outIPs now ... [%s]", name, cmdPat)
	outIPs := make(map[string]*FileIP)
	for oname, ofun := range outPathFuncs {
		opath := ofun(t)
		otgt := NewFileIP(opath)
		if outPortsDoStream[oname] {
			otgt.doStream = true
		}
		Debug.Printf("Task:%s: Creating outIP with path %s ...\n", name, otgt.Path())
		outIPs[oname] = otgt
	}
	t.OutIPs = outIPs
	t.Command = formatCommand(cmdPat, inIPs, outIPs, params, prepend)
	Debug.Printf("Task:%s: Created formatted command: %s [%s]", name, t.Command, cmdPat)
	return t
}

// InIP returns an IP for the in-port with name portName
func (t *Task) InIP(portName string) *FileIP {
	if t.InIPs[portName] == nil {
		Failf("No such in-portname (%s) in task (%s)\n", portName, t.Name)
	}
	return t.InIPs[portName]
}

// InPath returns the path name of an input file for the task
func (t *Task) InPath(portName string) string {
	return t.InIP(portName).Path()
}

// OutIP returns an IP for the in-port with name portName
func (t *Task) OutIP(portName string) *FileIP {
	if t.OutIPs[portName] == nil {
		Failf("No such out-portname (%s) in task (%s)\n", portName, t.Name)
	}
	return t.OutIPs[portName]
}

// OutPath returns the path name of an input file for the task
func (t *Task) OutPath(portName string) string {
	return t.OutIP(portName).Path()
}

// Param returns the value of a param, for the task
func (t *Task) Param(portName string) string {
	if param, ok := t.Params[portName]; ok {
		return param
	}
	Failf("No such param port '%s' for task '%s'\n", portName, t.Name)
	return "invalid"
}

// Execute executes the task (the shell command or go function configured for
// this task)
func (t *Task) Execute() {
	defer close(t.Done)

	if !t.anyOutputExists() && t.allFifosInOutIPsExist() {
		Debug.Printf("Task:%-12s Executing task. [%s]\n", t.Name, t.Command)

		// Create directories for out-IPs
		for _, oip := range t.OutIPs {
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
		finishTime := time.Now()
		execTime := finishTime.Sub(startTime)
		t.workflow.DecConcurrentTasks(t.cores)

		// Append audit info for the task to all its output IPs

		auditInfo := NewAuditInfo()
		auditInfo.Command = t.Command
		auditInfo.ProcessName = t.process.Name()
		auditInfo.Params = t.Params
		auditInfo.StartTime = startTime
		auditInfo.FinishTime = finishTime
		auditInfo.ExecTimeMS = execTime / time.Millisecond
		// Set the audit infos from incoming IPs into the "Upstream" map
		for _, iip := range t.InIPs {
			iipPath := iip.Path()
			iipAuditInfo := iip.AuditInfo()
			auditInfo.Upstream[iipPath] = iipAuditInfo
		}
		// Add the current audit info to output ips and write them to file
		for _, oip := range t.OutIPs {
			oip.SetAuditInfo(auditInfo)
			for _, iip := range t.InIPs {
				oip.AddKeys(iip.Keys())
			}
			oip.WriteAuditLogToFile()
		}

		Debug.Printf("Task:%-12s Atomizing IPs. [%s]\n", t.Name, t.Command)
		t.atomizeIPs()

	}
	Debug.Printf("Task:%s: Starting to send Done in t.Execute() ...) [%s]\n", t.Name, t.Command)
	t.Done <- 1
	Debug.Printf("Task:%s: Done sending Done, in t.Execute() [%s]\n", t.Name, t.Command)
}

// Check if any output file IP, or temporary file IPs, exist
func (t *Task) anyOutputExists() (anyFileExists bool) {
	anyFileExists = false
	for _, tgt := range t.OutIPs {
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
	for _, tgt := range t.OutIPs {
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
func (t *Task) allFifosInOutIPsExist() bool {
	for _, tgt := range t.OutIPs {
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
	for _, otgt := range t.OutIPs {
		if otgt.doStream {
			otgt.CreateFifo()
		}
	}
}

// Rename temporary output files to their proper file names
func (t *Task) atomizeIPs() {
	for _, tgt := range t.OutIPs {
		if !tgt.doStream {
			Debug.Printf("Atomizing file: %s -> %s", tgt.TempPath(), tgt.Path())
			tgt.Atomize()
			Debug.Printf("Done atomizing file: %s -> %s", tgt.TempPath(), tgt.Path())
		} else {
			Debug.Printf("IP is streaming, so not atomizing: %s", tgt.Path())
		}
	}
}

// Clean up any remaining FIFOs
// TODO: this is actually not really used anymore ...
func (t *Task) cleanUpFifos() {
	for _, tgt := range t.OutIPs {
		if tgt.doStream {
			Debug.Printf("Task:%s: Cleaning up FIFO for output IP: %s [%s]\n", t.Name, tgt.FifoPath(), t.Command)
			tgt.RemoveFifo()
		} else {
			Debug.Printf("Task:%s: output IP is not FIFO, so not removing any FIFO: %s [%s]\n", t.Name, tgt.Path(), t.Command)
		}
	}
}

var (
	trueVal  = true
	falseVal = false
)

func formatCommand(cmd string, inIPs map[string]*FileIP, outIPs map[string]*FileIP, params map[string]string, prepend string) string {
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
			if outIPs[name] == nil {
				msg := fmt.Sprint("Missing outpath for outport '", name, "' for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else {
				if typ == "o" {
					filePath = outIPs[name].TempPath() // Means important to Atomize afterwards!
				} else if typ == "os" {
					filePath = outIPs[name].FifoPath()
				}
			}
		} else if typ == "i" {
			// In-ports
			if inIPs[name] == nil {
				msg := fmt.Sprint("Missing in-IP for inport '", name, "' for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else if inIPs[name].Path() == "" && reduceInputs {
				ips := []*FileIP{}
				for ip := range inIPs[name].SubStream.Chan {
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
			} else if inIPs[name].Path() == "" {
				msg := fmt.Sprint("Missing inpath for inport '", name, "', and no substream, for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else {
				if inIPs[name].doStream {
					filePath = inIPs[name].FifoPath()
				} else {
					filePath = inIPs[name].Path()
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
