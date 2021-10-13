package scipipe

import (
	"crypto/sha1"
	"encoding/hex"
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
	Tags          map[string]string
	Done          chan int
	cores         int
	workflow      *Workflow
	Process       *Process
	portInfos     map[string]*PortInfo
	subStreamIPs  map[string][]*FileIP
}

// ------------------------------------------------------------------------
// Factory method(s)
// ------------------------------------------------------------------------

// NewTask instantiates and initializes a new Task
func NewTask(workflow *Workflow, process *Process, name string, cmdPat string, inIPs map[string]*FileIP, outPathFuncs map[string]func(*Task) string, portInfos map[string]*PortInfo, params map[string]string, tags map[string]string, prepend string, customExecute func(*Task), cores int) *Task {
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
		Process:       process,
		portInfos:     portInfos,
		subStreamIPs:  make(map[string][]*FileIP),
	}

	// Collect substream IPs
	for ptName, ptInfo := range portInfos {
		if ptInfo.join && ptInfo.joinSep != "" {
			// Merge multiple input paths from a substream on the IP, into one string
			ips := []*FileIP{}
			for ip := range inIPs[ptName].SubStream.Chan {
				ips = append(ips, ip)
			}
			t.subStreamIPs[ptName] = ips
		}
	}
	// Create Out-IPs
	for oname, outPathFunc := range outPathFuncs {
		oip, err := NewFileIP(outPathFunc(t))
		if err != nil {
			process.Fail(err.Error())
		}
		if ptInfo, ok := portInfos[oname]; ok {
			if ptInfo.doStream {
				oip.doStream = true
			}
		}
		t.OutIPs[oname] = oip
	}
	t.Command = t.formatCommand(cmdPat, portInfos, inIPs, t.subStreamIPs, t.OutIPs, params, tags, prepend)
	return t
}

const (
	parentDirPlaceHolder = "__parent__"
)

// formatCommand is a helper function for NewTask, that formats a shell command
// based on concrete file paths and parameter values
func (t *Task) formatCommand(cmd string, portInfos map[string]*PortInfo, inIPs map[string]*FileIP, subStreamIPs map[string][]*FileIP, outIPs map[string]*FileIP, params map[string]string, tags map[string]string, prepend string) string {
	r := getShellCommandPlaceHolderRegex()
	placeHolderMatches := r.FindAllStringSubmatch(cmd, -1)

	type placeHolderInfo struct {
		match     string
		portName  string
		modifiers []string
	}

	placeHolderInfos := make([]*placeHolderInfo, 0)
	for _, match := range placeHolderMatches {
		restMatch := match[2]
		parts := strings.Split(restMatch, "|")
		portName := parts[0]
		modifiers := parts[1:]
		placeHolderInfos = append(placeHolderInfos,
			&placeHolderInfo{
				portName:  portName,
				match:     match[0],
				modifiers: modifiers,
			})
	}

	for _, placeHolder := range placeHolderInfos {
		portName := placeHolder.portName
		portInfo := portInfos[portName]

		var replacement string
		switch portInfo.portType {

		case "o":
			if outIPs[portName] == nil {
				t.Failf("Missing outpath for outport (%s) for command (%s)", portName, cmd)
			}
			replacement = outIPs[portName].TempPath()
			replacement = applyPathModifiers(replacement, placeHolder.modifiers)
			replacement = replaceParentDirsWithPlaceholder(replacement)

		case "os":
			if outIPs[portName] == nil {
				t.Failf("Missing outpath for outport (%s) for command (%s)", portName, cmd)
			}
			replacement = outIPs[portName].FifoPath()
			replacement = applyPathModifiers(replacement, placeHolder.modifiers)
			if !strInSlice("basename", placeHolder.modifiers) {
				replacement = prependParentDirPath(replacement)
			}

		case "i":
			if inIPs[portName] == nil {
				t.Failf("Missing in-IP for inport (%s) for command (%s)", portName, cmd)
			}
			if portInfo.join && portInfo.joinSep != "" {
				// Merge multiple input paths from a substream on the IP, into one string
				paths := []string{}
				for _, ip := range subStreamIPs[portName] {
					path := ip.Path()
					path = applyPathModifiers(path, placeHolder.modifiers)
					path = prependParentDirPath(path)
					paths = append(paths, path)
				}
				replacement = strings.Join(paths, portInfo.joinSep)
			} else {
				if inIPs[portName].Path() == "" {
					t.Failf("Missing inpath for inport (%s), and no substream, for command (%s)", portName, cmd)
				}
				if inIPs[portName].doStream {
					replacement = inIPs[portName].FifoPath()
				} else {
					replacement = inIPs[portName].Path()
				}
				replacement = applyPathModifiers(replacement, placeHolder.modifiers)
				if !strInSlice("basename", placeHolder.modifiers) {
					replacement = prependParentDirPath(replacement)
				}
			}

		case "p":
			if params[portName] == "" {
				t.Failf("Missing param value for param (%s) for command (%s)", portName, cmd)
			} else {
				replacement = params[portName]
				replacement = applyPathModifiers(replacement, placeHolder.modifiers)
			}

		case "t":
			if tags[portName] == "" {
				t.Failf("Missing tag value for tag (%s) for command (%s)", portName, cmd)
			} else {
				replacement = tags[portName]
				replacement = applyPathModifiers(replacement, placeHolder.modifiers)
			}

		default:
			t.Failf("Replace failed for port (%s) for command (%s)", portName, cmd)
		}

		cmd = strings.Replace(cmd, placeHolder.match, replacement, -1)
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
		t.Failf("No such in-portname (%s)", portName)
	}
	return t.InIPs[portName]
}

// InPath returns the path name of an input file for the task
func (t *Task) InPath(portName string) string {
	return t.InIP(portName).Path()
}

// OutIP returns an IP for the in-port with name portName
func (t *Task) OutIP(portName string) *FileIP {
	if ip, ok := t.OutIPs[portName]; ok {
		return ip
	}
	t.Failf("No such out-portname (%s)", portName)
	return nil
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
	t.Failf("No such param port (%s)", portName)
	return "invalid"
}

// Tag returns the value of a param, for the task
func (t *Task) Tag(tagName string) string {
	if tag, ok := t.Tags[tagName]; ok {
		return tag
	}
	t.Failf("No such tag (%s)", tagName)
	return "invalid"
}

// ------------------------------------------------------------------------
// Execute the task
// ------------------------------------------------------------------------

// Execute executes the task (the shell command or go function in CustomExecute)
func (t *Task) Execute() {
	defer close(t.Done)

	// Do some sanity checks
	if t.tempDirsExist() {
		t.Failf("Existing temp folders found, so existing. Clean up temporary folders (starting with %s) before restarting the workflow!", tempDirPrefix)
	}

	if t.anyOutputsExist() {
		t.Done <- 1
		return
	}

	// Execute task
	t.workflow.IncConcurrentTasks(t.cores) // Will block if max concurrent tasks is reached
	err := t.createDirs()                  // Create output directories needed for any outputs
	if err != nil {
		t.Fail(err)
	}
	startTime := time.Now()
	if t.CustomExecute != nil {
		outputsStr := ""
		for oipName, oip := range t.OutIPs {
			outputsStr += " " + oipName + ": " + oip.Path()
		}
		t.Auditf("Executing: Custom Go function with outputs: %s", outputsStr)
		t.CustomExecute(t)
		t.Auditf("Finished: Custom Go function with outputs: %s", outputsStr)
	} else {
		t.Auditf("Executing: %s", t.Command)
		t.executeCommand(t.Command)
		t.Auditf("Finished: %s", t.Command)
	}
	finishTime := time.Now()
	t.writeAuditLogs(startTime, finishTime)

	t.ensureAllOutputsExist()
	atomizeErr := t.atomizeIPs()
	if atomizeErr != nil {
		t.Fail(atomizeErr)
	}

	t.workflow.DecConcurrentTasks(t.cores)

	t.Done <- 1
}

// ------------------------------------------------------------------------
// Helper methods for the Execute method
// ------------------------------------------------------------------------

// anyTempFileExists checks if any temporary workflow files exist and if so, returns true
func (t *Task) tempDirsExist() bool {
	if _, err := os.Stat(t.TempDir()); os.IsNotExist(err) {
		return false
	}
	return true
}

// anyOutputsExist if any output file IP, or temporary file IPs, exist
func (t *Task) anyOutputsExist() (anyFileExists bool) {
	anyFileExists = false
	for _, oip := range t.OutIPs {
		if !oip.doStream {
			opath := oip.Path()
			if _, err := os.Stat(opath); err == nil {
				t.Auditf("Output file already exists, so skipping: %s", opath)
				anyFileExists = true
			}
		}
	}
	return
}

// createDirs creates directories for out-IPs of the task
func (t *Task) createDirs() error {
	os.MkdirAll(t.TempDir(), 0777)

	for _, oip := range t.OutIPs {
		oipDir := oip.TempDir() // This will create all out dirs, including the temp dir
		if oip.doStream {       // Temp dirs are not created for fifo files
			oipDir = filepath.Dir(oip.FifoPath())
		} else {
			oipDir = t.TempDir() + "/" + oipDir
		}
		err := os.MkdirAll(oipDir, 0777)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not create directory: %s: %s", oipDir, err))
		}
	}

	return nil
}

// executeCommand executes the shell command cmd via bash
func (t *Task) executeCommand(cmd string) {
	// cd into the task's tempdir, execute the command, and cd back
	out, err := exec.Command("bash", "-c", "cd "+t.TempDir()+" && "+cmd+" && cd ..").CombinedOutput()
	if err != nil {
		t.Failf("Command failed!\nCommand:\n%s\n\nOutput:\n%s\nOriginal error:%s", cmd, string(out), err)
	}
}

func (t *Task) writeAuditLogs(startTime time.Time, finishTime time.Time) {
	// Append audit info for the task to all its output IPs
	auditInfo := NewAuditInfo()
	auditInfo.Command = t.Command
	auditInfo.ProcessName = t.Process.Name()
	auditInfo.Params = t.Params
	auditInfo.StartTime = startTime
	auditInfo.FinishTime = finishTime
	auditInfo.ExecTimeNS = finishTime.Sub(startTime)
	// Set the audit infos from incoming IPs into the "Upstream" map
	for inpName, iip := range t.InIPs {
		if t.portInfos[inpName].join {
			for _, subIP := range t.subStreamIPs[inpName] {
				auditInfo.Upstream[subIP.Path()] = subIP.AuditInfo()
			}
			continue
		}
		auditInfo.Upstream[iip.Path()] = iip.AuditInfo()
	}
	// Add output paths generated for this task
	for oipName, oip := range t.OutIPs {
		auditInfo.OutFiles[oipName] = oip.Path()
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

func (t *Task) ensureAllOutputsExist() {
	for _, ip := range t.OutIPs {
		filePath := filepath.Join(t.TempDir(), ip.TempPath())
		if _, err := os.Stat(filePath); os.IsNotExist(err) && !ip.doStream {
			t.Failf("Missing output temp-file (%s) for ip with path (%s)", filePath, ip.Path())
		}
	}
}

func (t *Task) atomizeIPs() error {
	outIPs := []*FileIP{}
	for _, ip := range t.OutIPs {
		outIPs = append(outIPs, ip)
	}
	return AtomizeIPs(t.TempDir(), outIPs...)
}

func (t *Task) Auditf(msg string, parts ...interface{}) {
	t.Audit(fmt.Sprintf(msg+"\n", parts...))
}

func (t *Task) Audit(msg string) {
	Audit.Printf("[Process:%s] %s", t.Process.Name(), msg)
}

func (t *Task) Failf(msg string, parts ...interface{}) {
	t.Fail(fmt.Sprintf(msg+"\n", parts...))
}

func (t *Task) Fail(msg interface{}) {
	Failf("[Task:%s] %s", t.Process.Name(), msg)
}

// AtomizeIPs renames temporary output files/directories to their proper paths.
// It is called both from Task, and from Process that implement cutom execution
// schedule.
func AtomizeIPs(tempExecDir string, ips ...*FileIP) error {
	for _, oip := range ips {
		// Move paths for ports, to final destinations
		if !oip.doStream {
			tempPath := tempExecDir + "/" + oip.TempPath()
			newPath := oip.Path()
			Debug.Println("Moving OutIP path: ", tempPath, " -> ", newPath)
			renameErr := os.Rename(tempPath, newPath)
			if renameErr != nil {
				return errors.New(fmt.Sprintf("Could not rename out-IP file %s to %s: %s", tempPath, newPath, renameErr))
			}
		}
	}
	// For remaining paths in temporary execution dir, just move out of it
	filepath.Walk(tempExecDir, func(tempPath string, fileInfo os.FileInfo, err error) error {
		if !fileInfo.IsDir() {
			newPath := strings.Replace(tempPath, tempExecDir+"/", "", 1)
			newPath = strings.Replace(newPath, FSRootPlaceHolder+"/", "/", 1)
			newPath = replacePlaceholdersWithParentDirs(newPath)
			newPathDir := filepath.Dir(newPath)
			if _, err := os.Stat(newPathDir); os.IsNotExist(err) {
				os.MkdirAll(newPathDir, 0777)
			}
			Debug.Println("Moving remaining file path: ", tempPath, " -> ", newPath)
			renameErr := os.Rename(tempPath, newPath)
			if renameErr != nil {
				return errors.New(fmt.Sprintf("Could not rename remaining file %s to %s: %s", tempPath, newPath, renameErr))
			}
		}
		return err
	})
	// Remove temporary execution dir (but not for absolute paths, or current dir)
	if tempExecDir != "" && tempExecDir != "." && tempExecDir[0] != '/' {
		remErr := os.RemoveAll(tempExecDir)
		if remErr != nil {
			return errors.New(fmt.Sprintf("Could not remove temp dir: %s: %s", tempExecDir, remErr))
		}
	}
	return nil
}

var tempDirPrefix = "_scipipe_tmp"

// TempDir returns a string that is unique to a task, suitable for use
// in file paths. It is built up by merging all input filenames and parameter
// values that a task takes as input, joined with dots.
func (t *Task) TempDir() string {
	pathPrefix := tempDirPrefix + "." + sanitizePathFragment(t.Name)
	hashPcs := []string{t.Name}
	for _, ipName := range sortedFileIPMapKeys(t.InIPs) {
		hashPcs = append(hashPcs, splitAllPaths(t.InIP(ipName).Path())...)
	}
	for _, subIPName := range sortedFileIPSliceMapKeys(t.subStreamIPs) {
		for _, subIPs := range t.subStreamIPs[subIPName] {
			hashPcs = append(hashPcs, splitAllPaths(subIPs.Path())...)
		}
	}
	for _, paramName := range sortedStringMapKeys(t.Params) {
		hashPcs = append(hashPcs, paramName+"_"+t.Param(paramName))
	}
	for _, tagName := range sortedStringMapKeys(t.Tags) {
		hashPcs = append(hashPcs, tagName+"_"+t.Tag(tagName))
	}

	// If resulting name is longer than 255
	if len(pathPrefix) > (255 - 40 - 1) {
		hashPcs = append(hashPcs, pathPrefix)
		pathPrefix = tempDirPrefix
	}
	sha1sum := sha1.Sum([]byte(strings.Join(hashPcs, "")))
	pathSegment := pathPrefix + "." + hex.EncodeToString(sha1sum[:])
	return pathSegment
}

func prependParentDirPath(path string) string {
	if path[0] == '/' {
		return path
	}
	// For relative paths, add ".." to get out of current dir
	return "../" + path
}

func replaceParentDirsWithPlaceholder(pathSegment string) string {
	return strings.ReplaceAll(pathSegment, "../", parentDirPlaceHolder)
}

func replacePlaceholdersWithParentDirs(pathSegment string) string {
	return strings.ReplaceAll(pathSegment, parentDirPlaceHolder, "../")
}
