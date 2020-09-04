package components

import (
	"fmt"
	"os"
	"testing"

	"log"

	"github.com/scipipe/scipipe"
)

func TestParamCombinator(t *testing.T) {
	os.MkdirAll(".tmp", 0744)

	var letters = []string{"a", "b"}
	var numbers = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"}

	// Create workflow
	wf := scipipe.NewWorkflow("wf", 4)

	lettersSource := NewParamSource(wf, "letter_source", letters...)
	numbersSource := NewParamSource(wf, "number_source", numbers...)

	paramCombiner := NewParamCombinator(wf, "param_combiner")
	paramCombiner.InParam("letters").From(lettersSource.Out())
	paramCombiner.InParam("numbers").From(numbersSource.Out())

	catenator := wf.NewProc("catenator", "echo {p:letters} {p:numbers} > {o:combined}")
	catenator.InParam("letters").From(paramCombiner.OutParam("letters"))
	catenator.InParam("numbers").From(paramCombiner.OutParam("numbers"))
	catenator.SetOut("combined", ".tmp/combined/{p:letters}.{p:numbers}.combined.txt")

	wf.Run()

	for _, l := range letters {
		for _, n := range numbers {
			filePath := fmt.Sprintf(".tmp/combined/%s.%s.combined.txt", l, n)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				log.Fatal("File did not exist: " + filePath)
			}
		}
	}

	// Clean up files
	filePaths := []string{}
	for _, l := range letters {
		for _, n := range numbers {
			filePaths = append(filePaths, fmt.Sprintf(".tmp/combined/%s.%s.combined.txt", l, n))
			filePaths = append(filePaths, filePaths[len(filePaths)-1]+".audit.json")
		}
	}
	for _, filePath := range filePaths {
		err := os.Remove(filePath)
		if err != nil {
			log.Fatal("Could not delete file:", filePath, "\n", err)
		}
	}
}
