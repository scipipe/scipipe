package scipipe

import (
	// "github.com/go-errors/errors"
	//"os"
	"math/rand"
	"os/exec"
	re "regexp"
	"time"
)

func ExecCmd(cmd string) string {
	Info.Println("Executing command: ", cmd)
	combOutput, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	if err != nil {
		Error.Fatalln("Could not execute command `" + cmd + "`: " + string(combOutput))
	}
	return string(combOutput)
}

func Check(err error, errMsg string) {
	if err != nil {
		Error.Println("Custom Error Message: " + errMsg)
		Error.Println("Original Error Message: " + err.Error())
		panic(err)
	}
}

func CheckErr(err error) {
	if err != nil {
		Error.Println(err)
		panic(err)
	}
}

// Return the regular expression used to parse the place-holder syntax for in-, out- and
// parameter ports, that can be used to instantiate a Process.
func getShellCommandPlaceHolderRegex() *re.Regexp {
	regex := "{(o|os|i|is|p):([^{}:]+)(:r(:([^{}:]))?)?}"
	r, err := re.Compile(regex)
	Check(err, "Could not compile regex: "+regex)
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
