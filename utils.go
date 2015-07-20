package scipipe

import (
	"fmt"
	"os/exec"
)

func ExecCmd(cmd string) {
	fmt.Println("Executing command: ", cmd)
	_, err := exec.Command("bash", "-lc", cmd).Output()
	Check(err)
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
