package scipipe

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Task represents a single static shell command, or go function, to be
// executed, and are scheduled and managed by a corresponding Process
type Task struct {
	Name          string
	Command       string
	CustomExecute func(*Task)
	InIPs         map[string]*FileIP
	OutIPs        map[string]*FileIP
	Params        map[string]string
	Done          chan int
	cores         int
	workflow      *Workflow
	process       *Process
}

// ------------------------------------------------------------------------
// Factory method(s)
// ------------------------------------------------------------------------

// NewTask instantiates and initializes a new Task
func NewTask(workflow *Workflow, process *Process, name string, cmdPat string, inIPs map[string]*FileIP, outPathFuncs map[string]func(*Task) string, outPortsDoStream map[string]bool, params map[string]string, prepend string, customExecute func(*Task), cores int) *Task {
	t := &Task{
		Name:          name,
		InIPs:         inIPs,
		OutIPs:        make(map[string]*FileIP),
		Params:        params,
		Command:       "",
		CustomExecute: customExecute,
		Done:          make(chan int),
		cores:         cores,
		workflow:      workflow,
		process:       process,
	}

	// Create Out-IPs
	for oname, outPathFunc := range outPathFuncs {
		oip := NewFileIP(outPathFunc(t))
		if outPortsDoStream[oname] {
			oip.doStream = true
		}
		t.OutIPs[oname] = oip
	}
	t.Command = formatCommand(cmdPat, inIPs, t.OutIPs, params, prepend)
	return t
}

// formatCommand is a helper function for NewTask, that formats a shell command
// based on concrete file paths and parameter values
func formatCommand(cmd string, inIPs map[string]*FileIP, outIPs map[string]*FileIP, params map[string]string, prepend string) string {
	r := getShellCommandPlaceHolderRegex()
	placeHolderMatches := r.FindAllStringSubmatch(cmd, -1)
	for _, placeHolderMatch := range placeHolderMatches {
		var filePath string

		placeHolderStr := placeHolderMatch[0]
		portType := placeHolderMatch[1]
		portName := placeHolderMatch[2]
		sep := " " // Default

		switch portType {
		case "o":
			if outIPs[portName] == nil {
				Fail("Missing outpath for outport '", portName, "' for command '", cmd, "'")
			}
			filePath = outIPs[portName].TempPath()
		case "os":
			if outIPs[portName] == nil {
				Fail("Missing outpath for outport '", portName, "' for command '", cmd, "'")
			}
			filePath = outIPs[portName].FifoPath()
		case "i":
			if inIPs[portName] == nil {
				Fail("Missing in-IP for inport '", portName, "' for command '", cmd, "'")
			}
			// Identify if the place holder represents a reduce-type in-port
			reduceInputs := false
			if len(placeHolderMatch) > 3 {
				sep = placeHolderMatch[5]
				reduceInputs = true
			}
			if reduceInputs && inIPs[portName].Path() == "" {
				// Merge multiple input paths from a substream on the IP, into one string
				ips := []*FileIP{}
				for ip := range inIPs[portName].SubStream.Chan {
					ips = append(ips, ip)
				}
				paths := []string{}
				for _, ip := range ips {
					paths = append(paths, ip.Path())
				}
				filePath = strings.Join(paths, sep)
			} else {
				if inIPs[portName].Path() == "" {
					Fail("Missing inpath for inport '", portName, "', and no substream, for command '", cmd, "'")
				}
				if inIPs[portName].doStream {
					filePath = inIPs[portName].FifoPath()
				} else {
					filePath = inIPs[portName].Path()
				}
			}
		case "p":
			if params[portName] == "" {
				msg := fmt.Sprint("Missing param value param '", portName, "' for command '", cmd, "'")
				CheckWithMsg(errors.New(msg), msg)
			} else {
				filePath = params[portName]
			}
		default:
			Fail("Replace failed for port ", portName, " for command '", cmd, "'")
		}
		cmd = strings.Replace(cmd, placeHolderStr, filePath, -1)
	}

	// Add prepend string to the command
	if prepend != "" {
		cmd = fmt.Sprintf("%s %s", prepend, cmd)
	}

	return cmd
}

// ------------------------------------------------------------------------
// Main API methods: Accessor methods
// ------------------------------------------------------------------------

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

// ------------------------------------------------------------------------
// Execute the task
// ------------------------------------------------------------------------

// Execute executes the task (the shell command or go function in CustomExecute)
func (t *Task) Execute() {
	defer close(t.Done)

	// Do some sanity checks
	if t.anyTempfileExists() {
		Failf("| %-32s | Existing temp files found so existing. Clean up .tmp files before restarting the workflow!", t.Name)
	}

	if t.anyOutputsExist() {
		t.Done <- 1
		return
	}

	// Execute task
	t.workflow.IncConcurrentTasks(t.cores) // Will block if max concurrent tasks is reached
	t.createDirs()                         // Create output directories needed for any outputs
	startTime := time.Now()
	if t.CustomExecute != nil {
		outputsStr := ""
		for oipName, oip := range t.OutIPs {
			outputsStr += " " + oipName + ": " + oip.Path()
		}
		Audit.Printf("| %-32s | Executing: Custom Go function with outputs: %s\n", t.Name, outputsStr)
		t.CustomExecute(t)
		Audit.Printf("| %-32s | Finished:  Custom Go function with outputs: %s\n", t.Name, outputsStr)
	} else {
		Audit.Printf("| %-32s | Executing: %s\n", t.Name, t.Command)
		t.executeCommand(t.Command)
		Audit.Printf("| %-32s | Finished:  %s\n", t.Name, t.Command)
	}
	finishTime := time.Now()
	t.writeAuditLogs(startTime, finishTime)
	t.atomizeIPs()
	t.workflow.DecConcurrentTasks(t.cores)

	t.Done <- 1
}

// ------------------------------------------------------------------------
// Helper methods for the Execute method
// ------------------------------------------------------------------------

// anyTempFileExists checks if any temporary workflow files exist and if so, returns true
func (t *Task) anyTempfileExists() (anyTempfileExists bool) {
	anyTempfileExists = false
	for _, oip := range t.OutIPs {
		if !oip.doStream {
			otmpPath := oip.TempPath()
			if _, err := os.Stat(otmpPath); err == nil {
				Warning.Printf("| %-32s | Temp file already exists: %s (Note: If resuming from a failed run, clean up .tmp files first. Also, make sure that two processes don't produce the same output files!).\n", t.Name, otmpPath)
				anyTempfileExists = true
			}
		}
	}
	return
}

// anyOutputsExist if any output file IP, or temporary file IPs, exist
func (t *Task) anyOutputsExist() (anyFileExists bool) {
	anyFileExists = false
	for _, oip := range t.OutIPs {
		if !oip.doStream {
			opath := oip.Path()
			if _, err := os.Stat(opath); err == nil {
				Audit.Printf("| %-32s | Output file already exists, so skipping: %s\n", t.Name, opath)
				anyFileExists = true
			}
		}
	}
	return
}

// createDirs creates directories for out-IPs of the task
func (t *Task) createDirs() {
	for _, oip := range t.OutIPs {
		oipDir := filepath.Dir(oip.Path())
		err := os.MkdirAll(oipDir, 0777)
		CheckWithMsg(err, "Could not create directory: "+oipDir)
	}

}

// executeCommand executes the shell command cmd via bash
func (t *Task) executeCommand(cmd string) {
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		Failf("Command failed!\nCommand:\n%s\n\nOutput:\n%s\nOriginal error:%s\n", cmd, string(out), err.Error())
	}
}

func (t *Task) writeAuditLogs(startTime time.Time, finishTime time.Time) {
	// Append audit info for the task to all its output IPs
	auditInfo := NewAuditInfo()
	auditInfo.Command = t.Command
	auditInfo.ProcessName = t.process.Name()
	auditInfo.Params = t.Params
	auditInfo.StartTime = startTime
	auditInfo.FinishTime = finishTime
	auditInfo.ExecTimeMS = finishTime.Sub(startTime) / time.Millisecond
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
}

// Rename temporary output files to their proper file names
func (t *Task) atomizeIPs() {
	for _, oip := range t.OutIPs {
		if !oip.doStream {
			oip.Atomize()
		}
	}
}
