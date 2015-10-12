// Example showing how you can wrap shell tasks in "static" tasks (with proper struct fields)

package main

import (
	"fmt"
	sci "github.com/samuell/scipipe"
)

func main() {
	fmt.Println("Starting ....")
	f := NewFooer()
	go f.Run()
	<-f.OutFoo
	fmt.Println("Done!")
}

// Components

type Fooer struct {
	InnerProc *sci.ShellProcess
	OutFoo    chan *sci.FileTarget
}

func NewFooer() *Fooer {
	innerFoo := sci.Shell("echo foo > {o:foo}")
	innerFoo.SetPathGenString("foo", "foo.txt")
	return &Fooer{
		InnerProc: innerFoo,
		OutFoo:    innerFoo.OutPorts["foo"],
	}
}

func (p *Fooer) Run() {
	p.InnerProc.Run()
}
