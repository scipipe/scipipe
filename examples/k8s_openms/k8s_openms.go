package main

import sp "github.com/scipipe/scipipe"

const (
	workDir = "/scipipe-data/"
)

func main() {
	prun := sp.NewPipelineRunner()

	peakPicker := sp.NewFromShell("PeakPicker", "PeakPickerHiRes -in {i:sample} -out {o:out} -ini {p:ini}")
	peakPicker.PathFormatters["out"] = func(t *sp.SciTask) string {
		//         filename = basename("{0}_{2}.{1}".format(*self.sampleFile.rsplit('.', 1) + ["peaks"]))
		//         return luigi.LocalTarget("results/"+filename)
		return "todo_implement_path_formatting.txt"
	}

	prun.AddProcess(peakPicker)

	prun.Run()
}
