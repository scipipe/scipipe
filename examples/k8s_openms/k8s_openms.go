package main

import (
	"path/filepath"
	str "strings"

	sp "github.com/scipipe/scipipe"
)

const (
	workDir = "/scipipe-data/"
)

func main() {
	prun := sp.NewPipelineRunner()

	sampleFilesSender := sp.NewIPQueue(workDir + "002_CRa_H9M5_M470_Pool_01_alternate_neg_low_mr.mzML")
	prun.AddProcess(sampleFilesSender)

	peakPicker := sp.NewFromShell("PeakPicker", "PeakPickerHiRes -in {i:sample} -out {o:out} -ini "+workDir+"openms-params/PPparam.ini")
	peakPicker.PathFormatters["out"] = func(t *sp.SciTask) string {
		parts := str.Split(filepath.Base(t.GetInPath("sample")), ".")
		outPath := workDir + "results/" + str.Join(parts[:len(parts)-1], "_") + ".peaks"
		return outPath
	}
	peakPicker.ExecMode = sp.ExecModeK8s
	peakPicker.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(peakPicker)

	sink := sp.NewSink()
	prun.AddProcess(sink)

	peakPicker.In["sample"].Connect(sampleFilesSender.Out)
	sink.Connect(peakPicker.Out["out"])

	prun.Run()
}
