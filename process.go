package scipipe

import (
	"errors"
	"os/user"
	str "strings"
)

// ExecMode specifies which execution mode should be used for a SciProcess and
// its corresponding SciTasks
type ExecMode int

const (
	// ExecModeLocal indicates that commands on the local computer
	ExecModeLocal ExecMode = iota
	// ExecModeSLURM indicates that commands should be executed on a HPC cluster
	// via a SLURM resource manager
	ExecModeSLURM ExecMode = iota
	// ExecModeK8s indicates that commands should be executed on a Kubernetes
	// cluster
	ExecModeK8s ExecMode = iota
)

type K8sConfMode int

const (
	K8sConfModeInCluster    K8sConfMode = iota
	K8sConfModeOutOfCluster K8sConfMode = iota
)

// ================== Process ==================

// Base interface for all processes
type Process interface {
	IsConnected() bool // Sanity check, to see whether all ports are connected
	Run()
}

// ================== SciProcess ==================

var DefaultImage = "perl"
var DefaultDataFolder = "/scipipe-data"

type SciProcess struct {
	Process
	Name             string
	CommandPattern   string
	ExecMode         ExecMode
	Prepend          string
	Spawn            bool
	In               map[string]*FilePort
	Out              map[string]*FilePort
	OutPortsDoStream map[string]bool
	PathFormatters   map[string]func(*SciTask) string
	ParamPorts       map[string]*ParamPort
	CustomExecute    func(*SciTask)
	KubeConfigPath   string
	K8sConfMode      K8sConfMode
	Image            string
	DataFolder       string
}

func NewSciProcess(name string, command string) *SciProcess {
	usr, err := user.Current()
	CheckErr(err)
	defaultKubeConfigPath := usr.HomeDir + "/.kube/config"
	Debug.Printf("Process %s: Setting default kube config path to %s\n", name, defaultKubeConfigPath)

	return &SciProcess{
		Name:             name,
		CommandPattern:   command,
		In:               make(map[string]*FilePort),
		Out:              make(map[string]*FilePort),
		OutPortsDoStream: make(map[string]bool),
		PathFormatters:   make(map[string]func(*SciTask) string),
		ParamPorts:       make(map[string]*ParamPort),
		Spawn:            true,
		KubeConfigPath:   defaultKubeConfigPath,
		Image:            DefaultImage,
		DataFolder:       DefaultDataFolder,
	}
}

// ----------- Main API init methods ------------

func NewFromShell(name string, cmd string) *SciProcess {
	if !LogExists {
		InitLogAudit()
	}
	p := NewSciProcess(name, cmd)
	p.initPortsFromCmdPattern(cmd, nil)
	return p
}

func ShellExpand(name string, cmd string, inPaths map[string]string, outPaths map[string]string, params map[string]string) *SciProcess {
	cmdExpr := expandCommandParamsAndPaths(cmd, params, inPaths, outPaths)
	p := NewSciProcess(name, cmdExpr)
	p.initPortsFromCmdPattern(cmdExpr, params)
	return p
}

func (p *SciProcess) GetInPort(portName string) *FilePort {
	if p.In[portName] != nil {
		return p.In[portName]
	} else {
		Error.Printf("No such in-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.Name)
	}
	return nil
}

func (p *SciProcess) GetOutPort(portName string) *FilePort {
	if p.Out[portName] != nil {
		return p.Out[portName]
	} else {
		Error.Printf("No such out-port ('%s') for process '%s'. Please check your workflow code!\n", portName, p.Name)
	}
	return nil
}

// ----------- Path formatting methods ------------

// SetPathStatic creates an (output) path formatter returning a static string file name
func (p *SciProcess) SetPathStatic(outPortName string, path string) {
	p.PathFormatters[outPortName] = func(t *SciTask) string {
		return path
	}
}

// SetPathExtend creates an (output) path formatter that extends the path of
// an input InformationPacket
func (p *SciProcess) SetPathExtend(inPortName string, outPortName string, extension string) {
	p.PathFormatters[outPortName] = func(t *SciTask) string {
		return t.InTargets[inPortName].GetPath() + extension
	}
}

// SetPathReplace creates an (output) path formatter that uses an input's path
// but replaces parts of it.
func (p *SciProcess) SetPathReplace(inPortName string, outPortName string, old string, new string) {
	p.PathFormatters[outPortName] = func(t *SciTask) string {
		return str.Replace(t.InTargets[inPortName].GetPath(), old, new, -1)
	}
}

// ------- Helper methods for initialization -------

func expandCommandParamsAndPaths(cmd string, params map[string]string, inPaths map[string]string, outPaths map[string]string) (cmdExpr string) {
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	if params != nil {
		Debug.Println("Params:", params)
	}
	if inPaths != nil {
		Debug.Println("inPaths:", inPaths)
	}
	if outPaths != nil {
		Debug.Println("outPaths:", outPaths)
	}
	cmdExpr = cmd
	Debug.Println("Got command: ", cmd)
	for _, m := range ms {
		placeHolderStr := m[0]
		typ := m[1]
		name := m[2]
		var filePath string
		if typ == "p" {
			if params != nil {
				if val, ok := params[name]; ok {
					Debug.Println("Found param:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = str.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		} else if typ == "i" {
			if inPaths != nil {
				if val, ok := inPaths[name]; ok {
					Debug.Println("Found inPath:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = str.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		} else if typ == "o" || typ == "os" {
			if outPaths != nil {
				if val, ok := outPaths[name]; ok {
					Debug.Println("Found outPath:", val)
					filePath = val
					Debug.Println("Replacing:", placeHolderStr, "->", filePath)
					cmdExpr = str.Replace(cmdExpr, placeHolderStr, filePath, -1)
				}
			}
		}
	}
	if cmd != cmdExpr {
		Debug.Printf("Expanded command '%s' into '%s'\n", cmd, cmdExpr)
	}
	return
}

// Set up in- and out-ports based on the shell command pattern used to create the
// SciProcess. Ports are set up in this way:
// `{i:PORTNAME}` specifies an in-port
// `{o:PORTNAME}` specifies an out-port
// `{os:PORTNAME}` specifies an out-port that streams via a FIFO file
// `{p:PORTNAME}` a "parameter-port", which means a port where parameters can be "streamed"
func (p *SciProcess) initPortsFromCmdPattern(cmd string, params map[string]string) {

	// Find in/out port names and Params and set up in struct fields
	r := getShellCommandPlaceHolderRegex()
	ms := r.FindAllStringSubmatch(cmd, -1)
	Debug.Printf("Got the following matches for placeholders in the command, when initializing ports: %v\n", ms)

	for _, m := range ms {
		if len(m) < 3 {
			msg := "Too few matches"
			Check(errors.New(msg), msg)
		}

		typ := m[1]
		name := m[2]
		if typ == "o" || typ == "os" {
			p.Out[name] = NewFilePort()
			if typ == "os" {
				p.OutPortsDoStream[name] = true
			}
		} else if typ == "i" {
			// Set up a channel on the inports, even though this is
			// often replaced by another processes output port channel.
			// It might be nice to have it init'ed with a channel
			// anyways, for use cases when we want to send InformationPacket
			// on the inport manually.
			p.In[name] = NewFilePort()
		} else if typ == "p" {
			if params == nil || params[name] == "" {
				p.ParamPorts[name] = NewParamPort()
			}
		}
	}
}

// ------- Sanity checks -------
func (proc *SciProcess) IsConnected() (isConnected bool) {
	isConnected = true
	for portName, port := range proc.In {
		if !port.IsConnected() {
			Error.Printf("InPort %s of process %s is not connected - check your workflow code!\n", portName, proc.Name)
			isConnected = false
		}
	}
	for portName, port := range proc.Out {
		if !port.IsConnected() {
			Error.Printf("OutPort %s of process %s is not connected - check your workflow code!\n", portName, proc.Name)
			isConnected = false
		}
	}
	for portName, port := range proc.ParamPorts {
		if !port.IsConnected() {
			Error.Printf("ParamPort %s of process %s is not connected - check your workflow code!\n", portName, proc.Name)
			isConnected = false
		}
	}
	return isConnected
}

// ============== SciProcess Run Method ===============

// Run runs the process by instantiating and executing SciTasks for all inputs
// and parameter values on its in-ports. in the case when there are no inputs
// or parameter values on the in-ports, it will run just once before it
// terminates. note that the actual execution of shell commands are done inside
// SciTask.Execute, not here.
func (p *SciProcess) Run() {
	defer p.closeOutPorts()

	tasks := []*SciTask{}
	Debug.Printf("Process %s: Starting to create and schedule tasks\n", p.Name)
	for t := range p.createTasks() {
		// Collect created tasks, for the second round
		// where tasks are waited for to finish, before
		// sending their outputs.
		Debug.Printf("Process %s: Instantiated task [%s] ...", p.Name, t.Command)
		tasks = append(tasks, t)

		anyPreviousFifosExists := t.anyFifosExist()

		if p.ExecMode == ExecModeLocal {
			if !anyPreviousFifosExists {
				Debug.Printf("Process %s: No FIFOs existed, so creating, for task [%s] ...", p.Name, t.Command)
				t.createFifos()
			}

			// Sending FIFOs for the task
			for oname, oip := range t.OutTargets {
				if oip.doStream {
					Debug.Printf("Process %s: Sending FIFO target on outport '%s' for task [%s] ...\n", p.Name, oname, t.Command)
					p.Out[oname].Chan <- oip
				}
			}
		}

		if anyPreviousFifosExists {
			Debug.Printf("Process %s: Previous FIFOs existed, so not executing task [%s] ...\n", p.Name, t.Command)
			// Since t.Execute() is not run, that normally sends the Done signal, we
			// have to send it manually here:
			go func() {
				defer close(t.Done)
				t.Done <- 1
			}()
		} else {
			Debug.Printf("Process %s: Go-Executing task in separate go-routine: [%s] ...\n", p.Name, t.Command)
			// Run the task
			go t.Execute()
			Debug.Printf("Process %s: Done go-executing task in go-routine: [%s] ...\n", p.Name, t.Command)
		}
	}

	Debug.Printf("Process %s: Starting to loop over %d tasks to send out targets ...\n", p.Name, len(tasks))
	for _, t := range tasks {
		Debug.Printf("Process %s: Waiting for Done from task: [%s]\n", p.Name, t.Command)
		<-t.Done
		Debug.Printf("Process %s: Received Done from task: [%s]\n", p.Name, t.Command)
		for oname, oip := range t.OutTargets {
			if !oip.doStream {
				Debug.Printf("Process %s: Sending target on outport %s, for task [%s] ...\n", p.Name, oname, t.Command)
				p.Out[oname].Chan <- oip
				Debug.Printf("Process %s: Done sending target on outport %s, for task [%s] ...\n", p.Name, oname, t.Command)
			}
		}
	}
}

// -------- Helper methods for the Run method ---------

func (p *SciProcess) receiveInputs() (inTargets map[string]*InformationPacket, inPortsOpen bool) {
	inPortsOpen = true
	inTargets = make(map[string]*InformationPacket)
	// Read input targets on in-ports and set up path mappings
	for inpName, inPort := range p.In {
		Debug.Printf("Process %s: Receieving on inPort %s ...", p.Name, inpName)
		inTarget, open := <-inPort.Chan
		if !open {
			inPortsOpen = false
			continue
		}
		Debug.Printf("Process %s: Got inTarget %s ...", p.Name, inTarget.GetPath())
		inTargets[inpName] = inTarget
	}
	return
}

func (p *SciProcess) receiveParams() (params map[string]string, paramPortsOpen bool) {
	paramPortsOpen = true
	params = make(map[string]string)
	// Read input targets on in-ports and set up path mappings
	for pname, pport := range p.ParamPorts {
		pval, open := <-pport.Chan
		if !open {
			paramPortsOpen = false
			continue
		}
		Debug.Println("Receiving param:", pname, "with value", pval)
		params[pname] = pval
	}
	return
}

func (p *SciProcess) createTasks() (ch chan *SciTask) {
	ch = make(chan *SciTask)
	go func() {
		defer close(ch)
		for {
			inTargets, inPortsOpen := p.receiveInputs()
			Debug.Printf("Process.createTasks:%s Got inTargets: %v", p.Name, inTargets)
			params, paramPortsOpen := p.receiveParams()
			Debug.Printf("Process.createTasks:%s Got params: %s", p.Name, params)
			if !inPortsOpen && !paramPortsOpen {
				Debug.Printf("Process.createTasks:%s Breaking: Both inPorts and paramPorts closed", p.Name)
				break
			}
			if len(p.In) == 0 && !paramPortsOpen {
				Debug.Printf("Process.createTasks:%s Breaking: No inports, and params closed", p.Name)
				break
			}
			if len(p.ParamPorts) == 0 && !inPortsOpen {
				Debug.Printf("Process.createTasks:%s Breaking: No params, and inPorts closed", p.Name)
				break
			}
			t := NewSciTask(p.Name, p.CommandPattern, inTargets, p.PathFormatters, p.OutPortsDoStream, params, p.Prepend, p.ExecMode)
			if p.CustomExecute != nil {
				t.CustomExecute = p.CustomExecute
			}
			t.KubeConfigPath = p.KubeConfigPath
			t.K8sConfMode = p.K8sConfMode
			t.Image = p.Image
			t.DataFolder = p.DataFolder
			ch <- t
			if len(p.In) == 0 && len(p.ParamPorts) == 0 {
				Debug.Printf("Process.createTasks:%s Breaking: No inports nor params", p.Name)
				break
			}
		}
	}()
	return ch
}

func (p *SciProcess) closeOutPorts() {
	for oname, oport := range p.Out {
		Debug.Printf("Process %s: Closing port %s ...\n", p.Name, oname)
		oport.Close()
	}
}
