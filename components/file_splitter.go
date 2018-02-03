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
	scipipe.EmptyWorkflowProcess
	name          string
	InFile        *scipipe.InPort
	OutSplitFile  *scipipe.OutPort
	LinesPerSplit int
	workflow      *scipipe.Workflow
}

// NewFileSplitter returns an initialized FileSplitter process that will split a
// file into multiple files, each with linesPerSplit number of lines per file
func NewFileSplitter(wf *scipipe.Workflow, name string, linesPerSplit int) *FileSplitter {
	fs := &FileSplitter{
		name:          name,
		InFile:        scipipe.NewInPort("in_file"),
		OutSplitFile:  scipipe.NewOutPort("out_split_file"),
		LinesPerSplit: linesPerSplit,
		workflow:      wf,
	}
	fs.InFile.Process = fs
	fs.OutSplitFile.Process = fs
	wf.AddProc(fs)
	return fs
}

// Name returns the name of the FileSplitter process
func (proc *FileSplitter) Name() string {
	return proc.name
}

// Run runs the FileSplitter process
func (proc *FileSplitter) Run() {
	defer proc.OutSplitFile.Close()

	rand.Seed(time.Now().UnixNano())

	fileReader := NewFileReader(proc.workflow, proc.Name()+"_file_reader"+getRandString(2))
	pop := scipipe.NewParamOutPort(proc.Name() + "_temp_filepath_feeder")
	pop.Process = proc
	fileReader.FilePath.Connect(pop)

	for ft := range proc.InFile.Chan {
		scipipe.Audit.Println("FileSplitter      Now processing input file ", ft.Path(), "...")

		go func() {
			defer pop.Close()
			pop.Send(ft.Path())
		}()

		pip := scipipe.NewParamInPort(proc.Name() + "temp_line_reader")
		pip.Process = proc
		pip.Connect(fileReader.OutLine)

		go fileReader.Run()

		i := 1
		splitIdx := 1
		splitFt := newSplitIPFromIndex(ft.Path(), splitIdx)
		if !splitFt.Exists() {
			splitfile := splitFt.OpenWriteTemp()
			for line := range pip.Chan {
				// If we have not yet reached the number of lines per split ...
				/// ... then just continue to write ...
				if i < splitIdx*proc.LinesPerSplit {
					splitfile.Write([]byte(line))
					i++
				} else {
					splitfile.Close()
					splitFt.Atomize()
					scipipe.Audit.Println("FileSplitter      Created split file", splitFt.Path())
					proc.OutSplitFile.Send(splitFt)
					splitIdx++

					splitFt = newSplitIPFromIndex(ft.Path(), splitIdx)
					splitfile = splitFt.OpenWriteTemp()
				}
			}
			splitfile.Close()
			splitFt.Atomize()
			scipipe.Audit.Println("FileSplitter      Created split file", splitFt.Path())
			proc.OutSplitFile.Send(splitFt)
		} else {
			scipipe.Audit.Printf("Split file already exists: %s, so skipping.\n", splitFt.Path())
		}
	}
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz")

func getRandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// IsConnected tells whether all the ports of the FileSplitter process are connected
func (proc *FileSplitter) IsConnected() bool {
	return proc.InFile.IsConnected() &&
		proc.OutSplitFile.IsConnected()
}

func newSplitIPFromIndex(basePath string, splitIdx int) *scipipe.IP {
	return scipipe.NewIP(basePath + fmt.Sprintf(".split_%v", splitIdx))
}
