package scipipe

import (
	// "github.com/go-errors/errors"
	//"os"
	"os/exec"
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
