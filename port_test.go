package scipipe

import (
	"os"
	"reflect"
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

	wf.Run()

	resultFiles := []string{"/tmp/hello_world.txt", "/tmp/tjena_world.txt"}

	for _, f := range resultFiles {
		_, err := os.Stat(f)
		if err != nil {
			t.Errorf("File not properly created: %s", f)
		}
	}

	cleanFiles(append(resultFiles, "/tmp/hello.txt", "/tmp/tjena.txt")...)
}

func TestInPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)

	inp := NewInPort("in_test")
	inp.process = NewProc(wf, "foo_proc", "echo foo > {o:out}")

	expectedName := "foo_proc.in_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of in-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}
}

func TestInPortSendRecv(t *testing.T) {
	inp := NewInPort("test_inport")
	inp.process = NewBogusProcess("bogus_process")

	ip := NewFileIP("/tmp/test.txt")
	go func() {
		inp.Send(ip)
	}()
	oip := inp.Recv()
	if ip != oip {
		t.Errorf("Received ip (with path %s) was not the same as the one sent (with path %s)", oip.Path(), ip.Path())
	}
}

func TestOutPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	outp := NewOutPort("out_test")
	outp.process = NewProc(wf, "foo_proc", "echo foo > {o:out}")

	expectedName := "foo_proc.out_test"
	if outp.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", outp.Name(), expectedName)
	}
}

func TestParamInPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	inp := NewParamInPort("in_test")
	inp.process = NewProc(wf, "foo_proc", "echo foo > {o:out}")

	expectedName := "foo_proc.in_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of in-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}
}

func TestParamInPortSendRecv(t *testing.T) {
	initTestLogs()

	pip := NewParamInPort("test_param_inport")
	param := "foo-bar"
	go func() {
		pip.Send(param)
	}()
	outParam := pip.Recv()
	if param != outParam {
		t.Errorf("Received param (%s) was not the same as the sent one (%s)", outParam, param)
	}
}

func TestParamInPortConnectStr(t *testing.T) {
	initTestLogs()

	pip := NewParamInPort("test_inport")
	pip.process = NewBogusProcess("bogus_process")

	pip.ConnectStr("foo", "bar", "baz")
	expectedStrs := []string{"foo", "bar", "baz"}

	outStrs := []string{}
	for s := range pip.Chan {
		outStrs = append(outStrs, s)
	}
	if !reflect.DeepEqual(outStrs, expectedStrs) {
		t.Errorf("Received strings %v are not the same as expected strings %v", outStrs, expectedStrs)
	}
}

func TestParamOutPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	pop := NewParamOutPort("out_test")
	pop.process = NewProc(wf, "foo_proc", "echo foo > {o:out}")

	expectedName := "foo_proc.out_test"
	if pop.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", pop.Name(), expectedName)
	}
}

func TestParamOutPortConnect(t *testing.T) {
	initTestLogs()

	popName := "test_param_outport"
	pop := NewParamOutPort(popName)
	pop.process = NewBogusProcess("bogus_process")

	pipName := "test_param_inport"
	pip := NewParamInPort(pipName)
	pip.process = NewBogusProcess("bogus_process")

	pop.Connect(pip)

	if !pop.Connected() {
		t.Errorf("Param out port '%s' not having connected status = true", pop.Name())
	}
	if !pip.Connected() {
		t.Errorf("Param out port '%s' not having connected status = true", pip.Name())
	}

	if pop.RemotePorts["bogus_process."+pipName] == nil {
		t.Errorf("ParamInPort not among remote ports in ParamOutPort")
	}
	if pip.RemotePorts["bogus_process."+popName] == nil {
		t.Errorf("ParamOutPort not among remote ports in ParamInPort")
	}
}
