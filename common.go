package scipipe

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	re "regexp"
	"strings"
	"time"

	"errors"
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
		err = errWrap(err, errMsg)
		Fail(err)
	}
}

// Check checks the error err, and prints the message in the error
func Check(err error) {
	if err != nil {
		Fail(err.Error())
	}
}

// Fail logs the error message, so that it will be possible to improve error
// messages in one place
func Fail(vs ...interface{}) {
	Error.Println(vs...)
	//Error.Println("Printing stack trace (read from bottom to find the workflow code that hit this error):")
	//debug.PrintStack()
	os.Exit(1) // Indicates a "general error" (See http://www.tldp.org/LDP/abs/html/exitcodes.html)
}

// Failf is like Fail but with msg being a formatter string for the message and
// vs being items to format into the message
func Failf(msg string, vs ...interface{}) {
	Fail(fmt.Sprintf(msg, vs...))
}

// Return the regular expression used to parse the place-holder syntax for in-, out- and
// parameter ports, that can be used to instantiate a Process.
func getShellCommandPlaceHolderRegex() *re.Regexp {
	regex := "{(o|os|i|is|p|t):([^{}]+)}"
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

func errWrap(err error, msg string) error {
	return errors.New(msg + "\nOriginal error: " + err.Error())
}

func errWrapf(err error, msg string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, v...) + "\nOriginal error: " + err.Error())
}

// splitAllPaths takes in a filepath and returns a list of all its parts
// e.g. "/a/b/c" becomes ["a" "b" "c'"]
func splitAllPaths(path string) []string {
	dir, file := filepath.Dir(path), filepath.Base(path)
	parts := []string{}
	for dir != file {
		parts = append([]string{file}, parts...)
		dir, file = filepath.Dir(dir), filepath.Base(dir)
	}
	return parts
}

// applyPathModifiers applies modifiers such as basename and %.ext for trimming
// file name extensions, which are added at the end of port-placeholders, separated
// by |-characters, such as: {i:infile|%.txt}, to get the path of "infile", but with the
// .txt extension trimmed.
// The full number of replacements are as follows:
// If the |-separated part looks like a search/replace
// pattern on the form of: s/SEARCHSTRING/REPLACESTRING/
// then execute the replacement.
// If the |-separated part starts iwth a %-character, then
// trim everything following that character, from the end
// of the path.
// If the |-separated part is "basename", then remove all leading
// folders up to the actual file name.
func applyPathModifiers(path string, modifiers []string) string {
	replacement := path

	substPtn := regexp.MustCompile("s\\/([^\\/]+)\\/([^\\/]*)\\/")
	trimEndPtn := regexp.MustCompile("%(.*)")
	basenamePtn := regexp.MustCompile(`.*\/`)
	dirnamePth := regexp.MustCompile(`\/[^\/]*$`)

	for _, modifier := range modifiers {
		// If the |-separated part looks like a search/replace
		// pattern on the form of: s/SEARCHSTRING/REPLACESTRING/
		// then execute the replacement.
		if substPtn.MatchString(modifier) {
			mbits := substPtn.FindStringSubmatch(modifier)
			search := mbits[1]
			replace := mbits[2]
			replacement = strings.Replace(replacement, search, replace, 1)
		}
		// If the |-separated part starts iwth a %-character, then
		// trim everything following that character, from the end
		// of the path.
		if trimEndPtn.MatchString(modifier) {
			mbits := trimEndPtn.FindStringSubmatch(modifier)
			end := mbits[1]
			if end == replacement[len(replacement)-len(end):] {
				replacement = replacement[:len(replacement)-len(end)]
			}
		}
		// If the |-separated part is "basename", then remove all leading
		// folders up to the actual file name.
		if modifier == "basename" {
			replacement = basenamePtn.ReplaceAllString(replacement, "")
		}

		// If the |-separated part is "dirname", then extract only
		// the directory of the path.
		if modifier == "dirname" {
			replacement = dirnamePth.ReplaceAllString(replacement, "")
		}
	}
	return replacement
}
