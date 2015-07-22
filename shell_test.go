package scipipe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	t "testing"
)

func TestShellHasInOutPorts(t *t.T) {
	tt := Sh("echo {i:in1} {o:out1}")
	tt.OutPathFuncs["out1"] = func() string {
		return fmt.Sprint(tt.InPaths["in1"], ".bar")
	}

	tt.InPorts["in1"] = make(chan *FileTarget, BUFSIZE)
	go func() { tt.InPorts["in1"] <- NewFileTarget("foo.txt") }()
	go tt.Run()
	<-tt.OutPorts["out1"]

	assert.NotNil(t, tt.InPorts["in1"], "InPorts are nil!")
	assert.NotNil(t, tt.OutPorts["out1"], "OutPorts are nil!")

	cleanFiles("foo.txt", "foo.txt.bar")
}

func TestShellCloseOutPortOnInPortClose(t *t.T) {
	fooTask := Sh("echo foo > {o:out1}")
	fooTask.OutPathFuncs["out1"] = func() string {
		return "foo.txt"
	}

	barReplacer := Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo") + ".bar"
	}

	barReplacer.InPorts["foo"] = fooTask.OutPorts["out1"]

	go fooTask.Run()
	go barReplacer.Run()
	<-barReplacer.OutPorts["bar"]

	// Assert no more content coming on channels
	assert.Nil(t, <-fooTask.OutPorts["out1"])
	assert.Nil(t, <-barReplacer.OutPorts["bar"])

	_, fooErr := os.Stat("foo.txt")
	assert.Nil(t, fooErr)
	_, barErr := os.Stat("foo.txt.bar")
	assert.Nil(t, barErr)

	cleanFiles("foo.txt", "foo.txt.bar")
}

func TestReplacePlaceholdersInCmd(t *t.T) {
	rawCmd := "echo {i:in1} > {o:out1}"
	tt := Sh(rawCmd)
	tt.OutPathFuncs["out1"] = func() string {
		return fmt.Sprint(tt.InPaths["in1"], ".bar")
	}

	tt.InPorts["in1"] = make(chan *FileTarget, BUFSIZE)
	ift := NewFileTarget("foo.txt")
	go func() {
		defer close(tt.InPorts["in1"])
		tt.InPorts["in1"] <- ift
	}()

	// Assert inport is still open after first read
	inportsOpen := tt.receiveInputs()
	assert.Equal(t, true, inportsOpen)

	// Assert inport is closed after second read
	inportsOpen = tt.receiveInputs()
	assert.Equal(t, false, inportsOpen)

	// Assert InPath is correct
	assert.Equal(t, "foo.txt", tt.InPaths["in1"], "foo.txt")

	// Assert placeholders are correctly replaced in command
	cmd := tt.replacePlaceholdersInCmd(rawCmd)
	assert.EqualValues(t, "echo foo.txt > foo.txt.bar", cmd)
}

// func TestParameterCommand(t *t.T) {
// 	tt := Sh("echo {p:a} {p:b} {p:c} > {o:out}")
// 	tt.OutPathFuncs["out"] = func() string {
// 		return fmt.Sprintf(
// 			"%s_%s_%s.txt",
// 			tt.Params["a"],
// 			tt.Params["b"],
// 			tt.Params["c"],
// 		)
// 	}
// 	// Feed the task with multiple combinations
// 	go func() {
// 		for _, a := range SS("a1", "a2", "a3") {
// 			for _, b := range SS("b1", "b2", "b3") {
// 				for _, c := range SS("c1", "c2", "c3") {
// 					tt.Params["a"] <- a
// 					tt.Params["b"] <- b
// 					tt.Params["c"] <- c
// 				}
// 			}
// 		}
// 	}()
// 	tt.Run()
// }

func cleanFiles(fileNames ...string) {
	for _, fileName := range fileNames {
		os.Remove(fileName)
	}
}
