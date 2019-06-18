package components

import (
	"fmt"
	"os"
	"testing"

	"log"

	"github.com/scipipe/scipipe"
)

func TestMapToTags(t *testing.T) {
	var numbers = []string{"1", "2", "3", "4"}

	// Create workflow
	wf := scipipe.NewWorkflow("wf", 4)
	numbersSource := NewParamSource(wf, "number_source", numbers...)

	numberFiles := wf.NewProc("make_files", "echo {p:number} > {o:out}")
	numberFiles.InParam("number").From(numbersSource.Out())
	numberFiles.SetOut("out", "{p:number}.txt")

	tagger := NewMapToTags(wf, "tagger", func(ip *scipipe.FileIP) map[string]string {
		tags := make(map[string]string)
		tags["file"] = ip.Path()
		return tags
	})
	tagger.In().From(numberFiles.Out("out"))

	catenator := wf.NewProc("cat", "cat {i:numbers} ../{t:numbers.file}")
	catenator.In("numbers").From(tagger.Out())

	wf.Run()

	// Clean Files
	filePaths := []string{}
	for _, n := range numbers {
		filePaths = append(filePaths, fmt.Sprintf("%s.txt", n))
		filePaths = append(filePaths, filePaths[len(filePaths)-1]+".audit.json")
	}
	for _, filePath := range filePaths {
		err := os.Remove(filePath)
		if err != nil {
			log.Fatal("Could not delete file:", filePath, "\n", err)
		}
	}
	err := os.RemoveAll("log")
	if err != nil {
		log.Fatal("Could not remove log directory\n", err)
	}

}
