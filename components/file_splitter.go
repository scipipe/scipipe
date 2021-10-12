package components

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

	for inIP := range p.InFile().Chan {
		lineNo := 1
		splitNo := 1
		splitIP := p.newSplitIPFromIndex(inIP.Path(), splitNo)
		if !splitIP.Exists() {
			inFile, err := os.Open(inIP.Path())
			if err != nil {
				err = errWrapf(err, "[FileSplitter] Could not open file %s", inIP.Path())
				p.Fail(err)
			}
			defer inFile.Close()
			taskDir := "_scipipe_tmp_" + p.Name() + "." + filepath.Base(inIP.Path())
			_, splitFile := p.createNewSplitFile(splitIP, taskDir)

			scanner := bufio.NewScanner(inFile)
			for scanner.Scan() {
				line := scanner.Text() + "\n"
				splitFile.WriteString(line)
				if lineNo == splitNo*p.LinesPerSplit {
					splitFile.Close()
					scipipe.FinalizePaths(taskDir, splitIP)
					p.OutSplitFile().Send(splitIP)
					// Create new IP
					splitNo++
					splitIP = p.newSplitIPFromIndex(inIP.Path(), splitNo)
					_, splitFile = p.createNewSplitFile(splitIP, taskDir)
				}
				lineNo++
			}
			splitFile.Close()
			scipipe.FinalizePaths(taskDir, splitIP)
			p.OutSplitFile().Send(splitIP)

			if scanner.Err() != nil {
				err = errWrapf(scanner.Err(), "[FileSplitter] Error when scanning input file %s", inIP.Path())
				p.Fail(err)
			}
		} else {
			p.Auditf("Split file already exists: %s, so skipping.", splitIP.Path())
		}
	}
}

func (p *FileSplitter) createNewSplitFile(ip *scipipe.FileIP, basePath string) (tempDir string, tempFile *os.File) {
	tempPath := basePath + "/" + ip.TempPath()
	tempDir = filepath.Dir(tempPath)
	err := os.MkdirAll(tempDir, 0777)
	if err != nil {
		err = errWrapf(err, "[FileSplitter] Could not create dirs for file %s", tempPath)
		p.Fail(err)
	}
	tempFile, err = os.Create(tempPath)
	if err != nil {
		scipipe.CheckWithMsg(err, "Could not create temp file "+tempPath)
	}
	return
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

func (p *FileSplitter) newSplitIPFromIndex(basePath string, splitIdx int) *scipipe.FileIP {
	ip, err := scipipe.NewFileIP(basePath + fmt.Sprintf(".split_%v", splitIdx))
	if err != nil {
		p.Fail(err)
	}
	return ip
}
