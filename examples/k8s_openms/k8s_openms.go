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
	// sampleFile = luigi.Parameter()
	//
	// "FeatureFinderMetabo",
	//     "-in", "/work/" + self.input().path,
	//     "-out", "/work/" + self.output().path,
	//     "-ini", "/work/openms-params/FFparam.ini"
	//
	// def requires(self):
	//     return PeakPickerTask(sampleFile=self.sampleFile)
	//
	// def output(self):
	//     filename = basename("{0}.featureXML".format(*self.sampleFile.rsplit('.', 1)))
	//     return luigi.LocalTarget("results/"+filename)

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

	peakPicker.In["sample"].Connect(sampleFilesSender.Out)
	sink.Connect(peakPicker.Out["out"])

	prun.Run()
}
