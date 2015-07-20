package main

import (
	"fmt"
	sp "github.com/samuell/scipipe"
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

	pl := sp.NewPipeline()

	// Init processes
	hisay := NewHiSayer(pl)
	split := NewStringSplitter(pl)
	lower := NewLowerCaser(pl)
	upper := NewUpperCaser(pl)
	zippr := NewZipper(pl)
	s2byt := sp.NewStrToByte(pl)
	filew := sp.NewFileWriterFromPath(pl, "out2.txt")
	// prntr := NewPrinter(pl)

	// Network definition *** This is where to look! ***
	split.In = hisay.Out
	lower.In = split.OutLeft
	upper.In = split.OutRight
	zippr.In1 = lower.Out
	zippr.In2 = upper.Out
	s2byt.In = zippr.Out
	filew.In = s2byt.Out

	pl.Run()
}

// ======= HiSayer =======

type hiSayer struct {
	Out chan string
}

func NewHiSayer(pl *sp.Pipeline) *hiSayer {
	t := &hiSayer{Out: make(chan string, BUFSIZE)}
	pl.AddTask(t)
	return t
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

func NewStringSplitter(pl *sp.Pipeline) *stringSplitter {
	t := &stringSplitter{
		OutLeft:  make(chan string, BUFSIZE),
		OutRight: make(chan string, BUFSIZE),
	}
	pl.AddTask(t)
	return t
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

func NewLowerCaser(pl *sp.Pipeline) *lowerCaser {
	t := &lowerCaser{Out: make(chan string, BUFSIZE)}
	pl.AddTask(t)
	return t
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

func NewUpperCaser(pl *sp.Pipeline) *upperCaser {
	t := &upperCaser{Out: make(chan string, BUFSIZE)}
	pl.AddTask(t)
	return t
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

func NewZipper(pl *sp.Pipeline) *zipper {
	t := &zipper{Out: make(chan string, BUFSIZE)}
	pl.AddTask(t)
	return t
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

func NewPrinter(pl *sp.Pipeline) *printer {
	t := &printer{}
	pl.AddTask(t)
	return t
}

func (proc *printer) Run() {
	for s := range proc.In {
		fmt.Println(s)
	}
}

// ======= task interface =======

type task interface {
	Run()
}
