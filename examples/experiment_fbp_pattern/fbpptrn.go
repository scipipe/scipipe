package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
	"math"
	"runtime"
	"strings"
)

const (
	BUFSIZE = 16384
)

// ======= Main =======

func main() {
	numThreads := runtime.NumCPU() - 1
	fmt.Println("Starting ", numThreads, " threads ...")
	runtime.GOMAXPROCS(numThreads)

	wf := sp.NewWorkflow("fbpptrn_wf", 4)

	// Init processes
	hisay := NewHiSayer("hisay")
	split := NewStringSplitter("split")
	lower := NewLowerCaser("lower")
	upper := NewUpperCaser("upper")
	zippr := NewZipper("zippr")
	prntr := NewPrinter("printr")

	// Network definition *** This is where to look! ***
	split.In = hisay.Out
	lower.In = split.OutLeft
	upper.In = split.OutRight
	zippr.In1 = lower.Out
	zippr.In2 = upper.Out
	prntr.In = zippr.Out

	wf.AddProcs(hisay, split, lower, upper, zippr)
	wf.SetDriver(prntr)
	wf.Run()
}

// ======= HiSayer =======

type hiSayer struct {
	name string
	Out  chan string
}

func NewHiSayer(name string) *hiSayer {
	return &hiSayer{
		name: name,
		Out:  make(chan string, BUFSIZE),
	}
}

func (proc *hiSayer) Name() string {
	return proc.name
}

func (proc *hiSayer) Run() {
	defer close(proc.Out)
	for i := 1; i <= 1e3; i++ {
		proc.Out <- fmt.Sprintf("Hi for the %d:th time!", i)
	}
}

func (proc *hiSayer) IsConnected() bool { return true }

// ======= StringSplitter =======

type stringSplitter struct {
	name     string
	In       chan string
	OutLeft  chan string
	OutRight chan string
}

func NewStringSplitter(name string) *stringSplitter {
	return &stringSplitter{
		name:     name,
		OutLeft:  make(chan string, BUFSIZE),
		OutRight: make(chan string, BUFSIZE),
	}
}
func (proc *stringSplitter) Name() string {
	return proc.name
}

func (proc *stringSplitter) Run() {
	defer close(proc.OutLeft)
	defer close(proc.OutRight)
	for s := range proc.In {
		halfLen := int(math.Floor(float64(len(s)) / float64(2)))
		proc.OutLeft <- s[0:halfLen]
		proc.OutRight <- s[halfLen:]
	}
}

func (proc *stringSplitter) IsConnected() bool { return true }

// ======= LowerCaser =======

type lowerCaser struct {
	name string
	In   chan string
	Out  chan string
}

func NewLowerCaser(name string) *lowerCaser {
	return &lowerCaser{name: name, Out: make(chan string, BUFSIZE)}
}
func (proc *lowerCaser) Name() string {
	return proc.name
}

func (proc *lowerCaser) Run() {
	defer close(proc.Out)
	for s := range proc.In {
		proc.Out <- strings.ToLower(s)
	}
}

func (proc *lowerCaser) IsConnected() bool { return true }

// ======= UpperCaser =======

type upperCaser struct {
	name string
	In   chan string
	Out  chan string
}

func NewUpperCaser(name string) *upperCaser {
	return &upperCaser{name: name, Out: make(chan string, BUFSIZE)}
}
func (proc *upperCaser) Name() string {
	return proc.name
}

func (proc *upperCaser) Run() {
	defer close(proc.Out)
	for s := range proc.In {
		proc.Out <- strings.ToUpper(s)
	}
}

func (proc *upperCaser) IsConnected() bool { return true }

// ======= Merger =======

type zipper struct {
	name string
	In1  chan string
	In2  chan string
	Out  chan string
}

func NewZipper(name string) *zipper {
	return &zipper{name: name, Out: make(chan string, BUFSIZE)}
}
func (proc *zipper) Name() string {
	return proc.name
}

func (proc *zipper) Run() {
	defer close(proc.Out)
	for {
		s1, ok1 := <-proc.In1
		s2, ok2 := <-proc.In2
		if !ok1 && !ok2 {
			break
		}
		proc.Out <- fmt.Sprint(s1, s2)
	}
}

func (proc *zipper) IsConnected() bool { return true }

// ======= Printer =======

type printer struct {
	name string
	In   chan string
}

func NewPrinter(name string) *printer {
	return &printer{name: name}
}
func (proc *printer) Name() string {
	return proc.name
}

func (proc *printer) Run() {
	for s := range proc.In {
		fmt.Println(s)
	}
}

func (proc *printer) IsConnected() bool { return true }
