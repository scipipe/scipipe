package scipipe

import (
	"fmt"
	"strconv"
	"strings"
)

type AccumulatorInt struct {
	Process
	In          chan *FileTarget
	Out         chan *FileTarget
	Accumulator int
	OutPath     string
}

func NewAccumulatorInt(outPath string) *AccumulatorInt {
	return &AccumulatorInt{
		In:          make(chan *FileTarget, BUFSIZE),
		Out:         make(chan *FileTarget, BUFSIZE),
		Accumulator: 0,
		OutPath:     outPath,
	}
}

func (proc *AccumulatorInt) Run() {
	defer close(proc.Out)
	for ft := range proc.In {
		Audit.Println("Processing file target %s ...", ft.GetPath())
		val, err := strconv.Atoi(strings.TrimSpace(string(ft.Read())))
		Check(err)
		fmt.Println("Got value %d ...", val)
		proc.Accumulator += val
	}
	outFt := NewFileTarget(proc.OutPath)

	outVal := fmt.Sprintf("%d", proc.Accumulator)
	outFt.WriteTempFile([]byte(outVal))

	outFt.Atomize()
	proc.Out <- outFt
}
