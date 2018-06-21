package components

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/scipipe/scipipe"
)

// FileSplitter is a process that will split a file into multiple files, each
// with LinesPerSplit number of lines per file
type FileSplitter struct {
	scipipe.BaseProcess
	LinesPerSplit int
}

// NewFileSplitter returns an initialized FileSplitter process that will split a
// file into multiple files, each with linesPerSplit number of lines per file
func NewFileSplitter(wf *scipipe.Workflow, name string, linesPerSplit int) *FileSplitter {
	p := &FileSplitter{
		BaseProcess:   scipipe.NewBaseProcess(wf, name),
		LinesPerSplit: linesPerSplit,
	}
	p.InitInPort(p, "file")
	p.InitOutPort(p, "split_file")
	wf.AddProc(p)
	return p
}

// InFile returns the port for the input file
func (p *FileSplitter) InFile() *scipipe.InPort { return p.InPort("file") }

// OutSplitFile returns the resulting split (part) files generated0
func (p *FileSplitter) OutSplitFile() *scipipe.OutPort { return p.OutPort("split_file") }

// Run runs the FileSplitter process
func (p *FileSplitter) Run() {
	defer p.CloseAllOutPorts()

	fileReader := NewFileReader(p.Workflow(), p.Name()+"_filereader_"+getRandString(7))
	pop := scipipe.NewOutParamPort(p.Name() + "_temp_filepath_feeder")
	pop.SetProcess(p)
	fileReader.InFilePath().From(pop)

	for ft := range p.InFile().Chan {
		scipipe.Audit.Println("FileSplitter                          Now processing input file ", ft.Path(), "...")

		go func() {
			defer pop.Close()
			pop.Send(ft.Path())
		}()

		pip := scipipe.NewInParamPort(p.Name() + "temp_line_reader")
		pip.SetProcess(p)
		pip.From(fileReader.OutLine())

		go fileReader.Run()

		lineNo := 1
		splitIdx := 1
		splitIP := newSplitIPFromIndex(ft.Path(), splitIdx)
		if !splitIP.Exists() {
			splitfile := splitIP.OpenWriteTemp()
			for line := range pip.Chan {
				// If we have not yet reached the number of lines per split ...
				/// ... then just continue to write ...
				if lineNo < splitIdx*p.LinesPerSplit {
					splitfile.Write([]byte(line))
					lineNo++
				} else {
					splitfile.Close()
					splitIP.Atomize()
					scipipe.Audit.Println("FileSplitter      Created split file", splitIP.Path())
					p.OutSplitFile().Send(splitIP)
					splitIdx++

					splitIP = newSplitIPFromIndex(ft.Path(), splitIdx)
					splitfile = splitIP.OpenWriteTemp()
				}
			}
			splitfile.Close()
			splitIP.Atomize()
			scipipe.Audit.Println("FileSplitter      Created split file", splitIP.Path())
			p.OutSplitFile().Send(splitIP)
		} else {
			scipipe.Audit.Printf("Split file already exists: %s, so skipping.\n", splitIP.Path())
		}
	}
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz")

func getRandString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func newSplitIPFromIndex(basePath string, splitIdx int) *scipipe.FileIP {
	return scipipe.NewFileIP(basePath + fmt.Sprintf(".split_%v", splitIdx))
}
