package components

import (
	"fmt"
	"os"
	"testing"

	"log"

	"github.com/scipipe/scipipe"
)

var letters = []string{"a", "b"}
var numbers = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"}

func TestFileCombinator(t *testing.T) {

	// Create letter files
	for _, s := range letters {
		fName := "/tmp/letterfile_" + s + ".txt"
		f, err := os.Create(fName)
		if err != nil {
			log.Fatalf("File could not be created: %s\n", fName)
		}
		f.WriteString(s)
	}

	// Create number files
	for _, s := range numbers {
		fName := "/tmp/numberfile_" + s + ".txt"
		f, err := os.Create(fName)
		if err != nil {
			log.Fatalf("File could not be created: %s\n", fName)
		}
		f.WriteString(s)
	}

	// Create workflow
	wf := scipipe.NewWorkflow("wf", 4)

	letterGlobber := NewFileGlobber(wf, "letter_globber", "/tmp/letterfile_*.txt")
	numberGlobber := NewFileGlobber(wf, "number_globber", "/tmp/numberfile_*.txt")

	fileCombiner := NewFileCombinator(wf, "file_combiner")
	fileCombiner.In("letters").From(letterGlobber.Out())
	fileCombiner.In("numbers").From(numberGlobber.Out())

	catenator := wf.NewProc("catenator", "cat {i:letters} {i:numbers} > {o:combined}")
	catenator.In("letters").From(fileCombiner.Out("letters"))
	catenator.In("numbers").From(fileCombiner.Out("numbers"))
	catenator.SetOut("combined", "/tmp/combined/{i:letters|basename|%.txt}.{i:numbers|basename|%.txt}.combined.txt")

	wf.Run()

	for _, l := range letters {
		for _, n := range numbers {
			filePath := fmt.Sprintf("/tmp/combined/letterfile_%s.numberfile_%s.combined.txt", l, n)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				log.Fatal("File did not exist: " + filePath)
			}
		}
	}

	// Clean up files
	filePaths := []string{}
	for _, s := range letters {
		filePaths = append(filePaths, fmt.Sprintf("/tmp/letterfile_%s.txt", s))
	}
	for _, s := range numbers {
		filePaths = append(filePaths, fmt.Sprintf("/tmp/numberfile_%s.txt", s))
	}
	for _, l := range letters {
		for _, n := range numbers {
			filePaths = append(filePaths, fmt.Sprintf("/tmp/combined/letterfile_%s.numberfile_%s.combined.txt", l, n))
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
