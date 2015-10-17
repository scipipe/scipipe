package main

import (
	sci "github.com/samuell/scipipe"
	str "strings"
)

func main() {
	sci.InitLogDebug()

	foo := sci.Shell("{o:foo}")
	foo.PathGen["foo"] = func(t *sci.ShellTask) string {
		return "foo.txt"
	}
	foo.CustomExecute = func(t *sci.ShellTask) {
		outf := t.OutTargets["foo"]
		outf.Write([]byte("foo"))
	}

	f2b := sci.Shell("{i:foo}{o:bar}")
	f2b.PathGen["bar"] = func(t *sci.ShellTask) string {
		return t.InTargets["foo"].GetPath() + ".bar"
	}
	f2b.CustomExecute = func(t *sci.ShellTask) {
		text := t.InTargets["foo"].Read()
		newText := str.Replace(string(text), "foo", "bar", 1)
		t.OutTargets["bar"].Write([]byte(newText))
	}

	// Connect stuff
	f2b.InPorts["foo"] = foo.OutPorts["foo"]

	// Run Pipeline
	pln := sci.NewPipeline()
	pln.AddProcs(foo, f2b)
	pln.Run()
}
