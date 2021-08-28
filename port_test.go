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
	hello.SetOut("hellofile", ".tmp/hello.txt")

	tjena := wf.NewProc("write_tjena", "echo tjena > {o:tjenafile}")
	tjena.SetOut("tjenafile", ".tmp/tjena.txt")

	world := wf.NewProc("append_world", "echo $(cat {i:infile}) world > {o:worldfile}")
	world.SetOut("worldfile", "{i:infile|%.txt}_world.txt")
	world.In("infile").From(hello.Out("hellofile"))
	world.In("infile").From(tjena.Out("tjenafile"))

	wf.Run()

	resultFiles := []string{".tmp/hello_world.txt", ".tmp/tjena_world.txt"}

	for _, f := range resultFiles {
		_, err := os.Stat(f)
		if err != nil {
			t.Errorf("File not properly created: %s", f)
		}
	}

	cleanFiles(append(resultFiles, ".tmp/hello.txt", ".tmp/tjena.txt")...)
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

	ip, err := NewFileIP(".tmp/test.txt")
	Check(err)
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

func TestInParamPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	inp := NewInParamPort("in_test")
	inp.process = NewProc(wf, "foo_proc", "echo foo > {o:out}")

	expectedName := "foo_proc.in_test"
	if inp.Name() != expectedName {
		t.Errorf("Name of in-port (%s) is not the expected (%s)", inp.Name(), expectedName)
	}
}

func TestInParamPortSendRecv(t *testing.T) {
	initTestLogs()

	pip := NewInParamPort("test_param_inport")
	param := "foo-bar"
	go func() {
		pip.Send(param)
	}()
	outParam := pip.Recv()
	if param != outParam {
		t.Errorf("Received param (%s) was not the same as the sent one (%s)", outParam, param)
	}
}

func TestInParamPortFromStr(t *testing.T) {
	initTestLogs()

	pip := NewInParamPort("test_inport")
	pip.process = NewBogusProcess("bogus_process")

	pip.FromStr("foo", "bar", "baz")
	expectedStrs := []string{"foo", "bar", "baz"}

	outStrs := []string{}
	for s := range pip.Chan {
		outStrs = append(outStrs, s)
	}
	if !reflect.DeepEqual(outStrs, expectedStrs) {
		t.Errorf("Received strings %v are not the same as expected strings %v", outStrs, expectedStrs)
	}
}

func TestOutParamPortName(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("dummy_workflow", 1)
	pop := NewOutParamPort("out_test")
	pop.process = NewProc(wf, "foo_proc", "echo foo > {o:out}")

	expectedName := "foo_proc.out_test"
	if pop.Name() != expectedName {
		t.Errorf("Name of out-port (%s) is not the expected (%s)", pop.Name(), expectedName)
	}
}

func TestOutParamPortFrom(t *testing.T) {
	initTestLogs()

	popName := "test_param_outport"
	pop := NewOutParamPort(popName)
	pop.process = NewBogusProcess("bogus_process")

	pipName := "test_param_inport"
	pip := NewInParamPort(pipName)
	pip.process = NewBogusProcess("bogus_process")

	pop.To(pip)

	if !pop.Ready() {
		t.Errorf("Param out port '%s' not having connected status = true", pop.Name())
	}
	if !pip.Ready() {
		t.Errorf("Param out port '%s' not having connected status = true", pip.Name())
	}

	if pop.RemotePorts["bogus_process."+pipName] == nil {
		t.Errorf("InParamPort not among remote ports in OutParamPort")
	}
	if pip.RemotePorts["bogus_process."+popName] == nil {
		t.Errorf("OutParamPort not among remote ports in InParamPort")
	}
}

func TestConnectBackwards(t *testing.T) {
	initTestLogs()

	wf := NewWorkflow("TestConnectBackwards", 16)

	p1 := wf.NewProc("p1", "echo foo > {o:foo}")
	p1.SetOutFunc("foo", func(t *Task) string { return "foo.txt" })

	p2 := wf.NewProc("p2", "sed 's/foo/bar/g' {i:foo} > {o:bar}")
	p2.SetOutFunc("bar", func(t *Task) string { return t.InPath("foo") + ".bar.txt" })

	p1.Out("foo").To(p2.In("foo"))

	wf.Run()

	cleanFiles("foo.txt", "foo.txt.bar.txt")
}
