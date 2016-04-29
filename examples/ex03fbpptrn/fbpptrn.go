package main

import (
	"fmt"
	sci "github.com/scipipe/scipipe"
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

	pipeline := sci.NewPipelineRunner()

	// Init processes
	hisay := NewHiSayer()
	split := NewStringSplitter()
	lower := NewLowerCaser()
	upper := NewUpperCaser()
	zippr := NewZipper()
	prntr := NewPrinter()

	// Network definition *** This is where to look! ***
	split.In = hisay.Out
	lower.In = split.OutLeft
	upper.In = split.OutRight
	zippr.In1 = lower.Out
	zippr.In2 = upper.Out
	prntr.In = zippr.Out

	pipeline.AddProcs(hisay, split, lower, upper, zippr, prntr)
	pipeline.Run()
}

// ======= HiSayer =======

type hiSayer struct {
	Out chan string
}

func NewHiSayer() *hiSayer {
	return &hiSayer{Out: make(chan string, BUFSIZE)}
}

func (proc *hiSayer) Run() {
	defer close(proc.Out)
	for i := 1; i <= 1e6; i++ {
		proc.Out <- fmt.Sprintf("Hi for the %d:th time!", i)
	}
}

// ======= StringSplitter =======

type stringSplitter struct {
	In       chan string
	OutLeft  chan string
	OutRight chan string
}

func NewStringSplitter() *stringSplitter {
	return &stringSplitter{
		OutLeft:  make(chan string, BUFSIZE),
		OutRight: make(chan string, BUFSIZE),
	}
}

func (proc *stringSplitter) Run() {
	defer close(proc.OutLeft)
	defer close(proc.OutRight)
	for s := range proc.In {
		halfLen := int(math.Floor(float64(len(s)) / float64(2)))
		proc.OutLeft <- s[0:halfLen]
		proc.OutRight <- s[halfLen:len(s)]
	}
}

// ======= LowerCaser =======

type lowerCaser struct {
	In  chan string
	Out chan string
}

func NewLowerCaser() *lowerCaser {
	return &lowerCaser{Out: make(chan string, BUFSIZE)}
}

func (proc *lowerCaser) Run() {
	defer close(proc.Out)
	for s := range proc.In {
		proc.Out <- strings.ToLower(s)
	}
}

// ======= UpperCaser =======

type upperCaser struct {
	In  chan string
	Out chan string
}

func NewUpperCaser() *upperCaser {
	return &upperCaser{Out: make(chan string, BUFSIZE)}
}

func (proc *upperCaser) Run() {
	defer close(proc.Out)
	for s := range proc.In {
		proc.Out <- strings.ToUpper(s)
	}
}

// ======= Merger =======

type zipper struct {
	In1 chan string
	In2 chan string
	Out chan string
}

func NewZipper() *zipper {
	return &zipper{Out: make(chan string, BUFSIZE)}
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

// ======= Printer =======

type printer struct {
	In chan string
}

func NewPrinter() *printer {
	return &printer{}
}

func (proc *printer) Run() {
	for s := range proc.In {
		fmt.Println(s)
	}
}
