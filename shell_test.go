package scipipe

import (
	"github.com/stretchr/testify/assert"
	t "testing"
)

func testShellHasInOutPorts(t *t.T) {
	testTask := Sh("echo {i:in1} {o:out1}")
	testTask.Run()
	assert.NotNil(t, testTask.OutPorts["in1"], "InPorts not nil!")
	assert.NotNil(t, testTask.OutPorts["out1"], "OutPorts not nil!")
}

func testShellCloseOutPortOnInPortClose(t *t.T) {
	fooTask := Sh("echo foo > {o:out1}")
	fooTask.OutPathFuncs["out1"] = func() string {
		return "foo.txt"
	}

	barReplacer := Sh("sed 's/foo/bar/g' {i:foo} > {o:bar}")
	barReplacer.OutPathFuncs["bar"] = func() string {
		return barReplacer.GetInPath("foo") + ".bar"
	}

	go fooTask.Run()
	go barReplacer.Run()

	<-barReplacer.OutPorts["bar"]
	assert.Nil(t, barReplacer.OutPorts["bar"], "bar OutPort was not nil!")
}
