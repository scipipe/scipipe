package main

import (
	str "strings"

	sp "github.com/scipipe/scipipe"
)

const (
	workDir = "/scipipe-data/"
)

func main() {
	prun := sp.NewPipelineRunner()

	peakPicker := sp.NewFromShell("PeakPicker", "PeakPickerHiRes -in {i:sample} -out {o:out} -ini {p:ini}")
	peakPicker.PathFormatters["out"] = func(t *sp.SciTask) string {
		parts := str.Split(t.GetInPath("sample"), ".")
		outPath := "results/" + str.Join(parts[:len(parts)-1], "_") + ".peaks"
		return outPath
	}

	prun.AddProcess(peakPicker)

	prun.Run()
}
