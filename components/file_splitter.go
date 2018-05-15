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
	pop := scipipe.NewParamOutPort(p.Name() + "_temp_filepath_feeder")
	pop.SetProcess(p)
	fileReader.InFilePath().Connect(pop)

	for ft := range p.InFile().Chan {
		scipipe.Audit.Println("FileSplitter                          Now processing input file ", ft.Path(), "...")

		go func() {
			defer pop.Close()
			pop.Send(ft.Path())
		}()

		pip := scipipe.NewParamInPort(p.Name() + "temp_line_reader")
		pip.SetProcess(p)
		pip.Connect(fileReader.OutLine())

		go fileReader.Run()

		i := 1
		splitIdx := 1
		splitFt := newSplitIPFromIndex(ft.Path(), splitIdx)
		if !splitFt.Exists() {
			splitfile := splitFt.OpenWriteTemp()
			for line := range pip.Chan {
				// If we have not yet reached the number of lines per split ...
				/// ... then just continue to write ...
				if i < splitIdx*p.LinesPerSplit {
					splitfile.Write([]byte(line))
					i++
				} else {
					splitfile.Close()
					splitFt.Atomize()
					scipipe.Audit.Println("FileSplitter      Created split file", splitFt.Path())
					p.OutSplitFile().Send(splitFt)
					splitIdx++

					splitFt = newSplitIPFromIndex(ft.Path(), splitIdx)
					splitfile = splitFt.OpenWriteTemp()
				}
			}
			splitfile.Close()
			splitFt.Atomize()
			scipipe.Audit.Println("FileSplitter      Created split file", splitFt.Path())
			p.OutSplitFile().Send(splitFt)
		} else {
			scipipe.Audit.Printf("Split file already exists: %s, so skipping.\n", splitFt.Path())
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
