package scipipe

import (
	"os"
	"reflect"
	"testing"
)

func TestConnectTo(t *testing.T) {
	initTestLogs()

	outpName := "test_outport"
	outp := NewOutPort(outpName)
	inpName := "test_inport"
	inp := NewInPort(inpName)

	ConnectTo(outp, inp)
	if !outp.IsConnected() {
		t.Error("Out-port not connected")
	}
	if !inp.IsConnected() {
		t.Error("In-port not connected")
	}
}

func TestConnectFrom(t *testing.T) {
	initTestLogs()

	outpName := "test_outport"
	outp := NewOutPort(outpName)
	inpName := "test_inport"
	inp := NewInPort(inpName)

	ConnectFrom(inp, outp)
	if !outp.IsConnected() {
		t.Error("Out-port not connected")
	}
	if !inp.IsConnected() {
		t.Error("In-port not connected")
	}
}

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

	cleanFiles(append(resultFiles, "/tmp/hello.txt", "/tmp/tjena.txt")...)
}

func TestInPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	prc := NewProc(wf, "foo_proc", "echo foo > {o:out}")
	inp := NewInPort("in_test")

	expectedName := "in_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of in-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}

	inp.Process = prc

	expectedName = "foo_proc.in_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of in-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}
}

func TestInPortSendRecv(t *testing.T) {
	inp := NewInPort("test_inport")
	ip := NewIP("/tmp/test.txt")
	go func() {
		inp.Send(ip)
	}()
	oip := inp.Recv()
	if ip != oip {
		t.Errorf("Received ip (with path %s) was not the same as the one sent (with path %s)", oip.GetPath(), ip.GetPath())
	}
}

func TestOutPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	prc := NewProc(wf, "foo_proc", "echo foo > {o:out}")
	inp := NewOutPort("out_test")

	expectedName := "out_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}

	inp.Process = prc

	expectedName = "foo_proc.out_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}
}

func TestParamInPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	prc := NewProc(wf, "foo_proc", "echo foo > {o:out}")
	inp := NewParamInPort("in_test")

	expectedName := "in_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of in-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}

	inp.Process = prc

	expectedName = "foo_proc.in_test"
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
	prc := NewProc(wf, "foo_proc", "echo foo > {o:out}")
	inp := NewParamOutPort("out_test")

	expectedName := "out_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}

	inp.Process = prc

	expectedName = "foo_proc.out_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}
}

func TestParamOutPortConnect(t *testing.T) {
	initTestLogs()

	popName := "test_param_outport"
	pop := NewParamOutPort(popName)
	pipName := "test_param_inport"
	pip := NewParamInPort(pipName)

	pop.Connect(pip)

	if !pop.IsConnected() {
		t.Errorf("Param out port '%s' not having connected status = true", pop.Name())
	}
	if !pip.IsConnected() {
		t.Errorf("Param out port '%s' not having connected status = true", pip.Name())
	}

	if pop.RemotePorts[pipName] == nil {
		t.Errorf("ParamInPort not among remote ports in ParamOutPort")
	}
	if pip.RemotePorts[popName] == nil {
		t.Errorf("ParamOutPort not among remote ports in ParamInPort")
	}
}
