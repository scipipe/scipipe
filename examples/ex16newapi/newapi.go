package main

import (
	"bytes"
	"fmt"

	r "reflect"

	"strings"

	sci "github.com/samuell/scipipe"
)

func main() {
	p := NewFooToBarReplacer()
	fmt.Println("Process: ", p)
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
	procStructVal := r.ValueOf(procStruct).Elem()
	for i := 0; i < procStructVal.NumField(); i++ {
		structFieldName := procStructVal.Type().Field(i).Name
		structFieldType := procStructVal.Type().Field(i).Type
		exampleChan := make(chan *sci.FileTarget)
		if strings.HasPrefix(structFieldName, "In") && structFieldType == r.TypeOf(exampleChan) {
			fmt.Println("In-port:", structFieldName)
		} else if strings.HasPrefix(structFieldName, "Out") && structFieldType == r.TypeOf(exampleChan) {
			fmt.Println("Out-port:", structFieldName)
		}
	}
	return procStruct
}
