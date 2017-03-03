package main

import (
	"path/filepath"
	str "strings"

	sp "github.com/scipipe/scipipe"
	spcomp "github.com/scipipe/scipipe/components"
)

const (
	workDir = "/scipipe-data/"
)

func main() {
	sp.InitLogWarning()

	prun := sp.NewPipelineRunner()

	altNegLowMRFiles := spcomp.NewFileGlobber(workDir + "*alternate_neg_low_mr.mzML")
	prun.AddProcess(altNegLowMRFiles)

	// -------------------------------------------------------------------
	// Peak Picker Process
	// -------------------------------------------------------------------
	peakPicker := sp.NewFromShell("peakpicker", "PeakPickerHiRes -in {i:sample} -out {o:peaks} -ini "+workDir+"openms-params/PPparam.ini")
	peakPicker.PathFormatters["peaks"] = func(t *sp.SciTask) string {
		parts := str.Split(filepath.Base(t.GetInPath("sample")), ".")
		peaksPath := workDir + "results/" + str.Join(parts[:len(parts)-1], "_") + ".peaks"
		return peaksPath
	}
	peakPicker.ExecMode = sp.ExecModeK8s
	peakPicker.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(peakPicker)

	// -------------------------------------------------------------------
	// Feature Finder process
	// -------------------------------------------------------------------
	featFinder := sp.NewFromShell("featfinder", "FeatureFinderMetabo -in {i:peaks} -out {o:feats} -ini "+workDir+"openms-params/FFparam.ini")
	featFinder.PathFormatters["feats"] = func(t *sp.SciTask) string {
		featsPath := t.GetInPath("peaks") + ".featureXML"
		return featsPath
	}
	featFinder.ExecMode = sp.ExecModeK8s
	featFinder.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(featFinder)

	// -------------------------------------------------------------------
	// Feature Linker process
	// -------------------------------------------------------------------
	strToSubstr := spcomp.NewStreamToSubStream()
	prun.AddProcess(strToSubstr)

	featLinker := sp.NewFromShell("featlinker", "FeatureLinkerUnlabeledQT -in {i:feats:r: } -out {o:consensus} -ini "+workDir+"openms-params/FLparam.ini -threads 2")
	featLinker.SetPathStatic("consensus", workDir+"results/"+"linked.consensusXML")
	featLinker.ExecMode = sp.ExecModeK8s
	featLinker.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(featLinker)

	// -------------------------------------------------------------------
	// File Filter process
	// -------------------------------------------------------------------
	fileFilter := sp.NewFromShell("filefilter", "FileFilter -in {i:unfiltered} -out {o:filtered} -ini "+workDir+"openms-params/FileFparam.ini")
	fileFilter.SetPathReplace("unfiltered", "filtered", "linked", "linked_filtered")
	fileFilter.ExecMode = sp.ExecModeK8s
	fileFilter.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(fileFilter)

	// -------------------------------------------------------------------
	// Text Exporter process
	// -------------------------------------------------------------------
	//
	// "TextExporter",
	//     "-in", "/work/" + self.input().path,
	//     "-out", "/work/" + self.output().path,
	//     "-ini", "/work/openms-params/TEparam.ini"
	//
	// def requires(self):
	//     return FileFilterTask(groupSuffix=self.groupSuffix)
	//
	// def output(self):
	//     return luigi.LocalTarget("results/"+self.groupSuffix+".csv")

	sink := sp.NewSink()
	prun.AddProcess(sink)

	// -------------------------------------------------------------------
	// Connect network
	// -------------------------------------------------------------------
	peakPicker.GetInPort("sample").Connect(altNegLowMRFiles.Out)
	featFinder.GetInPort("peaks").Connect(peakPicker.GetOutPort("peaks"))
	strToSubstr.In.Connect(featFinder.GetOutPort("feats"))
	featLinker.GetInPort("feats").Connect(strToSubstr.OutSubStream)
	fileFilter.GetInPort("unfiltered").Connect(featLinker.GetOutPort("consensus"))
	sink.Connect(fileFilter.GetOutPort("filtered"))

	prun.Run()
}
