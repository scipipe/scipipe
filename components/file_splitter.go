package components

import (
	"fmt"
	"github.com/scipipe/scipipe"
)

// File splitter component

type FileSplitter struct {
	scipipe.Process
	name          string
	InFile        *scipipe.FilePort
	OutSplitFile  *scipipe.FilePort
	LinesPerSplit int
}

func NewFileSplitter(name string, linesPerSplit int) *FileSplitter {
	return &FileSplitter{
		name:          name,
		InFile:        scipipe.NewFilePort(),
		OutSplitFile:  scipipe.NewFilePort(),
		LinesPerSplit: linesPerSplit,
	}
}

func (proc *FileSplitter) Name() string {
	return proc.name
}

func (proc *FileSplitter) Run() {
	defer proc.OutSplitFile.Close()

	fileReader := NewFileReader()

	for ft := range proc.InFile.Chan {
		scipipe.Audit.Println("FileSplitter      Now processing input file ", ft.GetPath(), "...")

		go func() {
			defer close(fileReader.FilePath)
			fileReader.FilePath <- ft.GetPath()
		}()

		go fileReader.Run()

		i := 1
		splitIdx := 1
		splitFt := newSplitInformationPacketFromIndex(ft.GetPath(), splitIdx)
		if !splitFt.Exists() {
			splitfile := splitFt.OpenWriteTemp()
			for line := range fileReader.OutLine {
				// If we have not yet reached the number of lines per split ...
				/// ... then just continue to write ...
				if i < splitIdx*proc.LinesPerSplit {
					splitfile.Write(line)
					i++
				} else {
					splitfile.Close()
					splitFt.Atomize()
					scipipe.Audit.Println("FileSplitter      Created split file", splitFt.GetPath())
					proc.OutSplitFile.Chan <- splitFt
					splitIdx++

					splitFt = newSplitInformationPacketFromIndex(ft.GetPath(), splitIdx)
					splitfile = splitFt.OpenWriteTemp()
				}
			}
			splitfile.Close()
			splitFt.Atomize()
			scipipe.Audit.Println("FileSplitter      Created split file", splitFt.GetPath())
			proc.OutSplitFile.Chan <- splitFt
		} else {
			scipipe.Audit.Printf("Split file already exists: %s, so skipping.\n", splitFt.GetPath())
		}
	}
}

func (proc *FileSplitter) IsConnected() bool {
	return proc.InFile.IsConnected() &&
		proc.OutSplitFile.IsConnected()
}

func newSplitInformationPacketFromIndex(basePath string, splitIdx int) *scipipe.InformationPacket {
	return scipipe.NewInformationPacket(basePath + fmt.Sprintf(".split_%v", splitIdx))
}
