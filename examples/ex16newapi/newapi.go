package main

import (
	"bytes"
	"fmt"

	r "reflect"

	sci "github.com/samuell/scipipe"
)

func main() {
	p := NewFooToBarReplacer()
	fmt.Println(p)
}

type FooToBarReplacer struct {
	sci.Process
	InnerProcess *sci.Process
	InFoo        chan *sci.FileTarget
	OutBar       chan *sci.FileTarget
}

func (p *FooToBarReplacer) Run() {
	// ...
}

func NewFooToBarReplacer() interface{} {
	execFunc := func(task *sci.ShellTask) {
		indata := task.InTargets["foo"].Read()
		indataReplaced := bytes.Replace(indata, []byte("foo"), []byte("bar"), -1)
		task.OutTargets["bar"].WriteTempFile(indataReplaced)
	}
	pathFuncs := map[string]func(*sci.ShellTask) string{
		"bar": func(t *sci.ShellTask) string { return t.InTargets["foo"].GetPath() + ".bar.txt" },
	}
	return NewProcessFromStruct(&FooToBarReplacer{}, execFunc, pathFuncs)
}

func NewProcessFromStruct(procStruct interface{}, execFunc func(*sci.ShellTask), pathFuncs map[string]func(*sci.ShellTask) string) interface{} {
	// Get in-ports of struct
	psVal := r.ValueOf(procStruct).Elem()
	for i := 0; i < psVal.NumField(); i++ {
		fmt.Println(psVal.Type().Field(i).Name)
	}
	return procStruct
}
