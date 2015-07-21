package main

import (
	"bytes"
	"fmt"
	sci "github.com/samuell/scipipe"
	"runtime"
	"sync"
)

const (
	BUFSIZE = 16
)

func main() {
	runtime.GOMAXPROCS(4)
	wf := NewExampleWorkflow()

	go func() {
		defer close(wf.In)
		defer close(wf.AReplaceWith)
		defer close(wf.CReplaceWith)
		defer close(wf.GReplaceWith)
		defer close(wf.TReplaceWith)

		for fileNo := 1; fileNo <= 16; fileNo++ {
			inFile := sci.NewFileTarget(fmt.Sprintf("file%d.txt", fileNo))
			fmt.Println("Processing file:", inFile.GetPath(), " ...")
			for _, ar := range []string{"B", "D"} {
				for _, cr := range []string{"E", "F"} {
					for _, gr := range []string{"H", "I"} {
						for _, tr := range []string{"J", "K"} {
							wf.In <- inFile
							wf.AReplaceWith <- ar
							wf.CReplaceWith <- cr
							wf.GReplaceWith <- gr
							wf.TReplaceWith <- tr
						}
					}
				}
			}
		}
	}()

	wf.Run()

	fmt.Println("Finished program")
}

// ======= Workflow task =======

type ExampleWorkflow struct {
	In           chan *sci.FileTarget
	AReplaceWith chan string
	CReplaceWith chan string
	GReplaceWith chan string
	TReplaceWith chan string
}

func (wf *ExampleWorkflow) Run() {

	pl := sci.NewPipeline()

	repl := NewReplaceLetters()
	repl.In = wf.In
	repl.AReplaceWith = wf.AReplaceWith
	repl.CReplaceWith = wf.CReplaceWith
	repl.GReplaceWith = wf.GReplaceWith
	repl.TReplaceWith = wf.TReplaceWith

	pl.AddTask(repl)
	fmt.Println("Trying to run workflow...")
	pl.Run()
}

func NewExampleWorkflow() *ExampleWorkflow {
	return &ExampleWorkflow{
		In:           make(chan *sci.FileTarget, BUFSIZE),
		AReplaceWith: make(chan string, BUFSIZE),
		CReplaceWith: make(chan string, BUFSIZE),
		GReplaceWith: make(chan string, BUFSIZE),
		TReplaceWith: make(chan string, BUFSIZE),
	}
}

// ====== Tasks =======

type ReplaceLetters struct {
	In           chan *sci.FileTarget
	Out          chan *sci.FileTarget
	AReplaceWith chan string
	CReplaceWith chan string
	GReplaceWith chan string
	TReplaceWith chan string
}

func (proc *ReplaceLetters) Run() {
	defer close(proc.Out)
	wg := new(sync.WaitGroup)
	for {
		inFile, okIn := <-proc.In
		arepl, okA := <-proc.AReplaceWith
		crepl, okC := <-proc.CReplaceWith
		grepl, okG := <-proc.GReplaceWith
		trepl, okT := <-proc.TReplaceWith
		if !okIn || !okA || !okC || !okG || !okT {
			break
		}
		wg.Add(1)
		go func() {
			fmt.Println("Processing:", inFile.GetPath(), arepl, crepl, grepl, trepl)
			outFilePath := fmt.Sprint(inFile.GetPath(), ".Arw", arepl, "_Crw", crepl, "_Grw", grepl, "_Trw", trepl)
			outFile := sci.NewFileTarget(outFilePath)
			text := inFile.Read()
			text = bytes.Replace(text, []byte("A"), []byte(arepl), -1)
			text = bytes.Replace(text, []byte("C"), []byte(crepl), -1)
			text = bytes.Replace(text, []byte("G"), []byte(grepl), -1)
			text = bytes.Replace(text, []byte("T"), []byte(trepl), -1)
			outFile.Write(text)
			wg.Done()
		}()
	}
	wg.Wait()
}

func NewReplaceLetters() *ReplaceLetters {
	return &ReplaceLetters{
		Out: make(chan *sci.FileTarget, BUFSIZE),
	}
}
