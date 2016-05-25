package main

import (
	sci "github.com/scipipe/scipipe"
)

func main() {
	sci.InitLogAudit()

	fq := sci.NewFileQueue("hej1.txt", "hej2.txt", "hej3.txt")
	fw := sci.Shell("filewriter", "echo {i:in} > {o:out}")
	fw.PathFormatters["out"] = func(t *sci.SciTask) string { return t.GetInPath("in") }
	sn := sci.NewSink()

	fw.InPorts["in"] = fq.Out
	sn.In = fw.OutPorts["out"]

	pl := sci.NewPipelineRunner()
	pl.AddProcesses(fq, fw, sn)

	pl.Run()
}
