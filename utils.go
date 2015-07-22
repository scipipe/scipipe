package scipipe

import (
	"fmt"
	// "github.com/go-errors/errors"
	"os"
	"os/exec"
)

func ExecCmd(cmd string) {
	fmt.Println("Executing command: ", cmd)
	_, err := exec.Command("bash", "-lc", cmd).Output()
	Check(err)
}

func Check(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
