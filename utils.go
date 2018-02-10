package scipipe

import (
	"math/rand"
	"os/exec"
	re "regexp"
	"time"
)

// ExecCmd executes the command cmd, as a shell command via bash
func ExecCmd(cmd string) string {
	Info.Println("Executing command: ", cmd)
	combOutput, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	if err != nil {
		Error.Fatalln("Could not execute command `" + cmd + "`: " + string(combOutput))
	}
	return string(combOutput)
}

// CheckWithMsg checks the error err, and prints both the original error message, and a
// custom one provided in errMsg
func CheckWithMsg(err error, errMsg string) {
	if err != nil {
		Error.Fatalln("Custom Error Message: " + errMsg + "\n" + "Original Error Message: " + err.Error())
	}
}

// Check checks the error err, and prints the message in the error
func Check(err error) {
	if err != nil {
		Error.Fatalln(err)
	}
}

// Return the regular expression used to parse the place-holder syntax for in-, out- and
// parameter ports, that can be used to instantiate a Process.
func getShellCommandPlaceHolderRegex() *re.Regexp {
	regex := "{(o|os|i|is|p):([^{}:]+)(:r(:([^{}:]))?)?}"
	r, err := re.Compile(regex)
	CheckWithMsg(err, "Could not compile regex: "+regex)
	return r
}

var letters = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func randSeqLC(n int) string {
	aseed := rand.NewSource(time.Now().UnixNano())
	arand := rand.New(aseed)
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[arand.Intn(len(letters))]
	}
	return string(b)
}
