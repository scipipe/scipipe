package proclib

import (
	"fmt"
	"github.com/scipipe/scipipe"
	"strconv"
	"strings"
)

type AccumulatorInt struct {
	scipipe.Process
	In          *scipipe.InPort
	Out         *scipipe.OutPort
	Accumulator int
	OutPath     string
}

func NewAccumulatorInt(outPath string) *AccumulatorInt {
	return &AccumulatorInt{
		Accumulator: 0,
		OutPath:     outPath,
	}
}

func (proc *AccumulatorInt) Run() {
	defer close(proc.Out.Chan)
	for ft := range proc.In.Chan {
		Audit.Printf("Accumulator:      Processing file target %s ...\n", ft.GetPath())
		val, err := strconv.Atoi(strings.TrimSpace(string(ft.Read())))
		Check(err)
		Debug.Printf("Accumulator:      Got value %d ...\n", val)
		proc.Accumulator += val
	}
	outFt := scipipe.NewFileTarget(proc.OutPath)
	outVal := fmt.Sprintf("%d", proc.Accumulator)
	outFt.WriteTempFile([]byte(outVal))
	outFt.Atomize()
	proc.Out.Chan <- outFt
}
