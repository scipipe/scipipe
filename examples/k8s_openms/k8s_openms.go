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

	sink := sp.NewSink()
	groupSuffixes := []string{"alternate_neg_low_mr", "alternate_neg_high_mr"}

	for _, groupSuffix := range groupSuffixes {

		altNegLowMRFiles := spcomp.NewFileGlobber(workDir + "*" + groupSuffix + ".mzML")
		prun.AddProcess(altNegLowMRFiles)

		// -------------------------------------------------------------------
		// Peak Picker Process
		// -------------------------------------------------------------------
		peakPicker := sp.NewFromShell("peakpicker-"+str.Replace(groupSuffix, "_", "-", -1), "PeakPickerHiRes -in {i:sample} -out {o:peaks} -ini "+workDir+"openms-params/PPparam.ini")
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
		featFinder := sp.NewFromShell("featfinder-"+str.Replace(groupSuffix, "_", "-", -1), "FeatureFinderMetabo -in {i:peaks} -out {o:feats} -ini "+workDir+"openms-params/FFparam.ini")
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

		featLinker := sp.NewFromShell("featlinker-"+str.Replace(groupSuffix, "_", "-", -1), "FeatureLinkerUnlabeledQT -in {i:feats:r: } -out {o:consensus} -ini "+workDir+"openms-params/FLparam.ini -threads 2")
		featLinker.SetPathStatic("consensus", workDir+"results/linked_"+groupSuffix+".consensusXML")
		featLinker.ExecMode = sp.ExecModeK8s
		featLinker.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
		prun.AddProcess(featLinker)

		// -------------------------------------------------------------------
		// File Filter process
		// -------------------------------------------------------------------
		fileFilter := sp.NewFromShell("filefilter-"+str.Replace(groupSuffix, "_", "-", -1), "FileFilter -in {i:unfiltered} -out {o:filtered} -ini "+workDir+"openms-params/FileFparam.ini")
		fileFilter.SetPathReplace("unfiltered", "filtered", "linked", "linked_filtered")
		fileFilter.ExecMode = sp.ExecModeK8s
		fileFilter.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
		prun.AddProcess(fileFilter)

		// -------------------------------------------------------------------
		// Text Exporter process
		// -------------------------------------------------------------------
		textExporter := sp.NewFromShell("textexport-"+str.Replace(groupSuffix, "_", "-", -1), "TextExporter -in {i:consensus} -out {o:csv} -ini "+workDir+"openms-params/TEparam.ini")
		textExporter.SetPathExtend("consensus", "csv", ".csv")
		textExporter.ExecMode = sp.ExecModeK8s
		textExporter.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
		prun.AddProcess(textExporter)

		// -------------------------------------------------------------------
		// Connect network
		// -------------------------------------------------------------------
		peakPicker.GetInPort("sample").Connect(altNegLowMRFiles.Out)
		featFinder.GetInPort("peaks").Connect(peakPicker.GetOutPort("peaks"))
		strToSubstr.In.Connect(featFinder.GetOutPort("feats"))
		featLinker.GetInPort("feats").Connect(strToSubstr.OutSubStream)
		fileFilter.GetInPort("unfiltered").Connect(featLinker.GetOutPort("consensus"))
		textExporter.GetInPort("consensus").Connect(fileFilter.GetOutPort("filtered"))
		sink.Connect(textExporter.GetOutPort("csv"))
	}
	prun.AddProcess(sink)

	prun.Run()
}
