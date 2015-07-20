package main

import (
	"fmt"
	"math"
	"reflect"
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

	pipeline := NewPipeline()

	// Init processes
	hisay := NewHiSayer(pipeline)
	split := NewStringSplitter(pipeline)
	lower := NewLowerCaser(pipeline)
	upper := NewUpperCaser(pipeline)
	zippr := NewZipper(pipeline)
	prntr := NewPrinter(pipeline)

	// Network definition *** This is where to look! ***
	split.In = hisay.Out
	lower.In = split.OutLeft
	upper.In = split.OutRight
	zippr.In1 = lower.Out
	zippr.In2 = upper.Out
	prntr.In = zippr.Out

	pipeline.Run()
}

// ======= HiSayer =======

type hiSayer struct {
	Out chan string
}

func NewHiSayer(pl *pipeline) *hiSayer {
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

func NewStringSplitter(pl *pipeline) *stringSplitter {
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

func NewLowerCaser(pl *pipeline) *lowerCaser {
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

func NewUpperCaser(pl *pipeline) *upperCaser {
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

func NewZipper(pl *pipeline) *zipper {
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

func NewPrinter(pl *pipeline) *printer {
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

// ======= Pipeline =======

type pipeline struct {
	tasks []task
}

func NewPipeline() *pipeline {
	return &pipeline{}
}

func (pl *pipeline) AddTask(t task) {
	pl.tasks = append(pl.tasks, t)
}

func (pl *pipeline) PrintPipeline() {
	sep := "------------------------------------------------------------------------"
	fmt.Println(sep)
	for i, task := range pl.tasks {
		fmt.Printf("Task %d: %v\n", i, reflect.TypeOf(task))
	}
	fmt.Println(sep)
}

func (pl *pipeline) Run() {
	for i, task := range pl.tasks {
		if i < len(pl.tasks)-1 {
			go task.Run()
		} else {
			task.Run()
		}
	}
}
