package components

import (
	"fmt"
	"github.com/scipipe/scipipe"
	"strconv"
	"strings"
)

type AccumulatorInt struct {
	scipipe.Process
	In          *scipipe.FilePort
	Out         *scipipe.FilePort
	Accumulator int
	OutPath     string
}

func NewAccumulatorInt(wf *scipipe.Workflow, outPath string) *AccumulatorInt {
	acc := &AccumulatorInt{
		Accumulator: 0,
		OutPath:     outPath,
	}
	wf.AddProc(acc)
	return acc
}

func (proc *AccumulatorInt) Run() {
	defer proc.Out.Close()
	for ft := range proc.In.InChan {
		scipipe.Audit.Printf("Accumulator:      Processing file target %s ...\n", ft.GetPath())
		val, err := strconv.Atoi(strings.TrimSpace(string(ft.Read())))
		Check(err)
		scipipe.Debug.Printf("Accumulator:      Got value %d ...\n", val)
		proc.Accumulator += val
	}
	outFt := scipipe.NewInformationPacket(proc.OutPath)
	outVal := fmt.Sprintf("%d", proc.Accumulator)
	outFt.WriteTempFile([]byte(outVal))
	outFt.Atomize()
	proc.Out.Send(outFt)
}
