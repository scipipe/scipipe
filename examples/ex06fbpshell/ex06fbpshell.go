package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	foo2bar := NewBarReplacer()

	go func() {
		defer close(foo2bar.InFoo)
		for _, v := range []string{"foo1", "foo2", "foo3"} {
			foo2bar.InFoo <- sci.NewFileTarget(v + ".txt")
		}
	}()

	go foo2bar.Run()

	for f := range foo2bar.OutBar {
		fmt.Println("Finished processing", f.GetPath(), "...")
	}
}

// ======= BarReplacer ========

type BarReplacer struct {
	InFoo  chan *sci.FileTarget
	OutBar chan *sci.FileTarget
}

func NewBarReplacer() *BarReplacer {
	t := &BarReplacer{
		InFoo:  make(chan *sci.FileTarget, sci.BUFSIZE),
		OutBar: make(chan *sci.FileTarget, sci.BUFSIZE),
	}
	return t
}

func (proc *BarReplacer) Run() {
	defer close(proc.OutBar)
	for inFile := range proc.InFoo {
		outFile := sci.NewFileTarget(inFile.GetPath() + ".bar")

		if !outFile.Exists() {
			cmd := fmt.Sprintf("sed 's/foo/bar/g' %s > %s", inFile.GetPath(), outFile.GetTempPath())
			sci.ExecCmd(cmd)
			outFile.Atomize()
		} else {
			fmt.Println("Skipping existing file: ", outFile.GetPath())
		}

		proc.OutBar <- outFile
	}
}
