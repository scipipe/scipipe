package components

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/scipipe/scipipe"
)

// CommandToParams takes a shell command, runs it, and sens each of its files
// as parameters on its OutParam parameter port
type CommandToParams struct {
	scipipe.BaseProcess
	command string
}

// NewCommandToParams returns an initialized new CommandToParams
func NewCommandToParams(wf *scipipe.Workflow, name string, command string) *CommandToParams {
	p := &CommandToParams{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		command:     command,
	}
	p.InitOutParamPort(p, "param")
	wf.AddProc(p)
	return p
}

// OutParam returns an parameter out-port with lines of the files being read
func (p *CommandToParams) OutParam() *scipipe.OutParamPort { return p.OutParamPort("param") }

// Run the CommandToParams
func (p *CommandToParams) Run() {
	defer p.CloseAllOutPorts()

	out, err := exec.Command("bash", "-c", p.command).CombinedOutput()
	if err != nil {
		panic("Could not run command: " + p.command + "\nERROR: " + err.Error())
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		param := scanner.Text()
		p.OutParamPort("param").Send(param)
	}
}
