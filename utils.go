package scipipe

import (
	// "github.com/go-errors/errors"
	//"os"
	"os/exec"
	re "regexp"
)

func ExecCmd(cmd string) {
	Info.Println("Executing command: ", cmd)
	combOutput, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	if err != nil {
		Error.Println("Could not execute command `" + cmd + "`: " + string(combOutput))
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// Short-hand function to create a slice of strings
func SS(ss ...string) []string {
	sslice := []string{}
	for _, s := range ss {
		sslice = append(sslice, s)
	}
	return sslice
}

func copyMapStrStr(m map[string]string) (nm map[string]string) {
	nm = make(map[string]string)
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

// Return the regular expression used to parse the place-holder syntax for in-, out- and
// parameter ports, that can be used to instantiate a SciProcess.
func getShellCommandPlaceHolderRegex() *re.Regexp {
	r, err := re.Compile("{(o|os|i|is|p):([^{}:]+)}")
	Check(err)
	return r
}
