package scipipe

import (
	"github.com/stretchr/testify/assert"
	t "testing"
)

func testShellHasInOutPorts(t *t.T) {
	testTask := Sh("echo {i:in1} {o:out1}")
	testTask.Init()
	assert.NotNil(t, testTask.OutPorts["in1"], "InPorts not nil!")
	assert.NotNil(t, testTask.OutPorts["out1"], "OutPorts not nil!")
}
