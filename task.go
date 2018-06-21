package scipipe

import (
	"crypto/sha1"
	"encoding/hex"
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
	Tags          map[string]string
	Done          chan int
	cores         int
	workflow      *Workflow
	process       *Process
}

// ------------------------------------------------------------------------
// Factory method(s)
// ------------------------------------------------------------------------

// NewTask instantiates and initializes a new Task
func NewTask(workflow *Workflow, process *Process, name string, cmdPat string, inIPs map[string]*FileIP, outPathFuncs map[string]func(*Task) string, outPortsDoStream map[string]bool, params map[string]string, tags map[string]string, prepend string, customExecute func(*Task), cores int) *Task {
	t := &Task{
		Name:          name,
		InIPs:         inIPs,
		OutIPs:        make(map[string]*FileIP),
		Params:        params,
		Tags:          tags,
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
	t.Command = formatCommand(cmdPat, inIPs, t.OutIPs, params, tags, prepend)
	return t
}

// formatCommand is a helper function for NewTask, that formats a shell command
// based on concrete file paths and parameter values
func formatCommand(cmd string, inIPs map[string]*FileIP, outIPs map[string]*FileIP, params map[string]string, tags map[string]string, prepend string) string {
	r := getShellCommandPlaceHolderRegex()
	placeHolderMatches := r.FindAllStringSubmatch(cmd, -1)
	portInfos := map[string]map[string]string{}
	for _, placeHolderMatch := range placeHolderMatches {
		portName := placeHolderMatch[2]
		portInfos[portName] = map[string]string{}
		portInfos[portName]["match"] = placeHolderMatch[0]
		portInfos[portName]["port_type"] = placeHolderMatch[1]
		portInfos[portName]["port_name"] = portName
		portInfos[portName]["reduce_inputs"] = "false"
		// Identify if the place holder represents a reduce-type in-port
		if len(placeHolderMatch) > 5 {
			portInfos[portName]["reduce_inputs"] = "true"
			portInfos[portName]["sep"] = placeHolderMatch[7]
		}
	}

	for portName, portInfo := range portInfos {
		var filePath string
		switch portInfo["port_type"] {
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
			if portInfo["reduce_inputs"] == "true" && inIPs[portName].Path() == "" {
				// Merge multiple input paths from a substream on the IP, into one string
				ips := []*FileIP{}
				for ip := range inIPs[portName].SubStream.Chan {
					ips = append(ips, ip)
				}
				paths := []string{}
				for _, ip := range ips {
					paths = append(paths, parentDirPath(ip.Path()))
				}
				filePath = strings.Join(paths, portInfo["sep"])
			} else {
				if inIPs[portName].Path() == "" {
					Fail("Missing inpath for inport '", portName, "', and no substream, for command '", cmd, "'")
				}
				if inIPs[portName].doStream {
					filePath = parentDirPath(inIPs[portName].FifoPath())
				} else {
					filePath = parentDirPath(inIPs[portName].Path())
				}
			}
		case "p":
			if params[portName] == "" {
				msg := fmt.Sprint("Missing param value for param '", portName, "' for command '", cmd, "'")
				Fail(msg)
			} else {
				filePath = params[portName]
			}
		case "t":
			if tags[portName] == "" {
				msg := fmt.Sprint("Missing tag value for tag '", portName, "' for command '", cmd, "'")
				Fail(msg)
			} else {
				filePath = tags[portName]
			}
		default:
			Fail("Replace failed for port ", portName, " for command '", cmd, "'")
		}
		cmd = strings.Replace(cmd, portInfo["match"], filePath, -1)
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

// Tag returns the value of a param, for the task
func (t *Task) Tag(tagName string) string {
	if tag, ok := t.Tags[tagName]; ok {
		return tag
	}
	Failf("No such tag '%s' for task '%s'\n", tagName, t.Name)
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
		LogAuditf(t.Name, "Executing: Custom Go function with outputs: %s", outputsStr)
		t.CustomExecute(t)
		LogAuditf(t.Name, "Executing: Custom Go function with outputs: %s", outputsStr)
	} else {
		LogAuditf(t.Name, "Executing: %s", t.Command)
		t.executeCommand(t.Command)
		LogAuditf(t.Name, "Finished: %s", t.Command)
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
			otmpPath := t.TempDir() + "/" + oip.TempPath()
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
	os.MkdirAll(t.TempDir(), 0777)
	for _, oip := range t.OutIPs {
		oipDir := oip.TempDir() // This will create all out dirs, including the temp dir
		if oip.doStream {       // Temp dirs are not created for fifo files
			oipDir = filepath.Dir(oip.FifoPath())
		} else {
			oipDir = t.TempDir() + "/" + oipDir
		}
		err := os.MkdirAll(oipDir, 0777)
		CheckWithMsg(err, "Could not create directory: "+oipDir)
	}

}

// executeCommand executes the shell command cmd via bash
func (t *Task) executeCommand(cmd string) {
	// cd into the task's tempdir, execute the command, and cd back
	out, err := exec.Command("bash", "-c", "cd "+t.TempDir()+" && "+cmd+" && cd ..").CombinedOutput()
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
			oip.AddTags(iip.Tags())
		}
		oip.WriteAuditLogToFile()
	}
}

func (t *Task) atomizeIPs() {
	outIPs := []*FileIP{}
	for _, ip := range t.OutIPs {
		outIPs = append(outIPs, ip)
	}
	AtomizeIPs(t.TempDir(), outIPs...)
}

// AtomizeIPs renames temporary output files/directories to their proper paths.
// It is called both from Task, and from Process that implement cutom execution
// schedule.
func AtomizeIPs(tempExecDir string, ips ...*FileIP) {
	for _, oip := range ips {
		// Move paths for ports, to final destinations
		if !oip.doStream {
			os.Rename(tempExecDir+"/"+oip.TempPath(), oip.Path())
		}
	}
	// For remaining paths in temporary execution dir, just move out of it
	filepath.Walk(tempExecDir, func(tempPath string, fileInfo os.FileInfo, err error) error {
		if !fileInfo.IsDir() {
			newPath := strings.Replace(tempPath, tempExecDir+"/", "", 1)
			newPath = strings.Replace(newPath, FSRootPlaceHolder+"/", "/", 1)
			newPathDir := filepath.Dir(newPath)
			if _, err := os.Stat(newPathDir); os.IsNotExist(err) {
				os.MkdirAll(newPathDir, 0777)
			}
			Debug.Println("Moving: ", tempPath, " -> ", newPath)
			renameErr := os.Rename(tempPath, newPath)
			CheckWithMsg(renameErr, "Could not rename file "+tempPath+" to "+newPath)
		}
		return err
	})
	// Remove temporary execution dir
	remErr := os.RemoveAll(tempExecDir)
	CheckWithMsg(remErr, "Could not remove temp dir: "+tempExecDir)
}

// TempDir returns a string that is unique to a task, suitable for use
// in file paths. It is built up by merging all input filenames and parameter
// values that a task takes as input, joined with dots.
func (t *Task) TempDir() string {
	pathPcs := []string{"tmp." + t.Name}
	for _, ipName := range sortedFileIPMapKeys(t.InIPs) {
		pathPcs = append(pathPcs, filepath.Base(t.InIP(ipName).Path()))
	}
	for _, paramName := range sortedStringMapKeys(t.Params) {
		pathPcs = append(pathPcs, paramName+"_"+t.Param(paramName))
	}
	for _, tagName := range sortedStringMapKeys(t.Tags) {
		pathPcs = append(pathPcs, tagName+"_"+t.Tag(tagName))
	}
	pathSegment := strings.Join(pathPcs, ".")
	if len(pathSegment) > 255 {
		sha1sum := sha1.Sum([]byte(pathSegment))
		pathSegment = t.Name + "." + hex.EncodeToString(sha1sum[:])
	}
	return pathSegment
}

func parentDirPath(path string) string {
	if path[0] == '/' {
		return path
	}
	// For relative paths, add ".." to get out of current dir
	return "../" + path
}
