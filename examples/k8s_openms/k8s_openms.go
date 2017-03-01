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
	prun := sp.NewPipelineRunner()

	sampleFilesSender := sp.NewIPQueue(workDir + "002_CRa_H9M5_M470_Pool_01_alternate_neg_low_mr.mzML")
	prun.AddProcess(sampleFilesSender)

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
		featsPath := t.GetInPath("peaks") + ".features.xml"
		return featsPath
	}
	featFinder.ExecMode = sp.ExecModeK8s
	featFinder.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(featFinder)

	// -------------------------------------------------------------------
	// Feature Linker process
	// -------------------------------------------------------------------
	// groupSuffix = luigi.Parameter()
	// inputList = map(lambda i: "/work/" + i.path, self.input())
	// inputStr = ' '.join(inputList)
	//
	// "command": ["sh","-c"],
	// "args": [
	//     "FeatureLinkerUnlabeledQT -in " + inputStr +
	//     " -out /work/" + self.output().path +
	//     " -ini /work/openms-params/FLparam.ini" +
	//     " -threads 2"],
	//
	// def requires(self):
	//     inputFiles = glob.glob("data/*_"+self.groupSuffix+".mzML")
	//     return map(lambda f: FeatureFinderTask(sampleFile=f),inputFiles)
	//
	// def output(self):
	//     return luigi.LocalTarget("results/linked_"+self.groupSuffix+".consensusXML")
	strToSubstr := spcomp.NewStreamToSubStream()
	prun.AddProcess(strToSubstr)

	featLinker := sp.NewFromShell("featlinker", "FeatureLinkerUnlabeledQT -in {i:feats:r: } -out {o:consensus} -ini "+workDir+"openms-params/FLparam.ini -threads 2")
	featLinker.PathFormatters["consensus"] = func(t *sp.SciTask) string {
		featsPath := t.GetInPath("feats") + ".consensus.xml"
		return featsPath
	}
	featLinker.ExecMode = sp.ExecModeK8s
	featLinker.Image = "container-registry.phenomenal-h2020.eu/phnmnl/openms:v1.11.1_cv0.1.9"
	prun.AddProcess(featLinker)

	// -------------------------------------------------------------------
	// File Filter process
	// -------------------------------------------------------------------
	// groupSuffix = luigi.Parameter()
	//
	// "FileFilter",
	//     "-in", "/work/" + self.input().path,
	//     "-out", "/work/" + self.output().path,
	//     "-ini", "/work/openms-params/FileFparam.ini"
	//
	// def requires(self):
	//     return FeatureLinkerTask(groupSuffix=self.groupSuffix)
	//
	// def output(self):
	//     return luigi.LocalTarget("results/linked_filtered_"+self.groupSuffix+".consensusXML")

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
	peakPicker.In["sample"].Connect(sampleFilesSender.Out)
	featFinder.In["peaks"].Connect(peakPicker.Out["peaks"])
	strToSubstr.In = featFinder.Out["feats"]
	featLinker.In["feats"] = strToSubstr.OutSubStream
	sink.Connect(featLinker.Out["consensus"])

	prun.Run()
}
