package scipipe

import (
	"os"
	"testing"
)

func TestMultiInPort(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("test_multiinport_wf", 4)
	hello := wf.NewProc("write_hello", "echo hello > {o:hellofile}")
	hello.SetPathStatic("hellofile", "/tmp/hello.txt")

	tjena := wf.NewProc("write_tjena", "echo tjena > {o:tjenafile}")
	tjena.SetPathStatic("tjenafile", "/tmp/tjena.txt")

	world := wf.NewProc("append_world", "echo $(cat {i:infile}) world > {o:worldfile}")
	world.SetPathReplace("infile", "worldfile", ".txt", "_world.txt")
	world.In("infile").Connect(hello.Out("hellofile"))
	world.In("infile").Connect(tjena.Out("tjenafile"))

	wf.ConnectLast(world.Out("worldfile"))
	wf.Run()

	resultFiles := []string{"/tmp/hello_world.txt", "/tmp/tjena_world.txt"}

	for _, f := range resultFiles {
		_, err := os.Stat(f)
		if err != nil {
			t.Errorf("File not properly created: %s", f)
		}
	}

	cleanFiles(resultFiles...)
}
