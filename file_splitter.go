package scipipe

import (
	"fmt"
)

// File splitter component

type FileSplitter struct {
	Process
	InFile        chan *FileTarget
	OutSplitFile  chan *FileTarget
	LinesPerSplit int
}

func NewFileSplitter(linesPerSplit int) *FileSplitter {
	return &FileSplitter{
		InFile:        make(chan *FileTarget, BUFSIZE),
		OutSplitFile:  make(chan *FileTarget, BUFSIZE),
		LinesPerSplit: linesPerSplit,
	}
}

func (proc *FileSplitter) Run() {
	defer close(proc.OutSplitFile)

	if !LogExists {
		InitLogAudit()
	}

	fileReader := NewFileReader()

	for ft := range proc.InFile {
		Audit.Println("FileSplitter      Now processing input file ", ft.GetPath(), "...")

		go func() {
			defer close(fileReader.FilePath)
			fileReader.FilePath <- ft.GetPath()
		}()

		go fileReader.Run()

		i := 1
		splitIdx := 1
		splitFt := newSplitFileTargetFromIndex(ft.GetPath(), splitIdx)
		if !splitFt.Exists() {
			splitfile := splitFt.OpenWriteTemp()
			for line := range fileReader.OutLine {
				if i < splitIdx*proc.LinesPerSplit {
					splitfile.Write(line)
					i++
				} else {
					splitfile.Close()
					splitFt.Atomize()
					Audit.Println("FileSplitter      Created split file", splitFt.GetPath())
					proc.OutSplitFile <- splitFt
					splitIdx++
					splitFt = newSplitFileTargetFromIndex(ft.GetPath(), splitIdx)
					splitfile = splitFt.OpenWriteTemp()
				}
			}
		} else {
			Audit.Printf("Split file already exists: %s, so skipping.\n", splitFt.GetPath())
		}
	}
}

func newSplitFileTargetFromIndex(basePath string, splitIdx int) *FileTarget {
	return NewFileTarget(basePath + fmt.Sprintf(".split_%v", splitIdx))
}
