package main

import (
	sci "github.com/samuell/scipipe"
)

func main() {
	sci.InitLogInfo()

	fq := sci.FQ("hej1.txt", "hej2.txt", "hej3.txt")
	fw := sci.Sh("echo hej > {i:in}")

	fw.InPorts["in"] = fq.Out

	pl := sci.NewPipeline()
	pl.AddTasks(fq, fw)
	pl.Run()
}
