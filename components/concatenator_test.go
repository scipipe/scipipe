package components

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/scipipe/scipipe"
)

var (
	tempDir   = ".tmp"
	testFiles = map[string]string{
		"a.txt": "A1",
		"a.csv": "A2",
		"b.txt": "B1",
		"b.csv": "B2",
		"b.tsv": "B3",
		"c.txt": "C1",
	}
)

func TestConcatenator(t *testing.T) {
	createTestFiles()

	// Create and run workflow
	wf := scipipe.NewWorkflow("wf", 4)
	fileSrc := NewFileSource(wf, "filesrc", tempDir+"/a.txt", tempDir+"/b.txt", tempDir+"/c.txt")
	concat := NewConcatenator(wf, "concat", tempDir+"/concatenated.txt")
	concat.In().From(fileSrc.Out())
	wf.Run()

	haveContentByte, err := ioutil.ReadFile(tempDir + "/concatenated.txt")
	if err != nil {
		panic(err)
	}
	haveContent := string(haveContentByte)

	wantContent := `A1
B1
C1
`

	if haveContent != wantContent {
		t.Fatalf("Wanted: %s but got %s", wantContent, haveContent)
	}

	cleanUpTestFiles()
}

func TestConcatenatorGroupByTag(t *testing.T) {
	createTestFiles()

	// Create and run workflow
	wf := scipipe.NewWorkflow("wf", 4)
	fileSrc := NewFileSource(wf, "filesrc",
		tempDir+"/a.txt",
		tempDir+"/a.csv",
		tempDir+"/b.txt",
		tempDir+"/b.csv",
		tempDir+"/b.tsv",
		tempDir+"/c.txt")

	dsNameAdder := NewMapToTags(wf, "dsname_adder", func(ip *scipipe.FileIP) map[string]string {
		dsFileName := filepath.Base(ip.Path())
		pat, err := regexp.Compile(`\.[^\.]*$`)
		if err != nil {
			panic(err)
		}
		datasetName := pat.ReplaceAllString(dsFileName, "")
		scipipe.Audit.Printf("Got dataset_name: %s", datasetName)
		return map[string]string{"dataset_name": datasetName}
	})
	dsNameAdder.In().From(fileSrc.Out())

	concat := NewConcatenator(wf, "concat", tempDir+"/concat.txt")
	concat.In().From(dsNameAdder.Out())
	concat.GroupByTag = "dataset_name"
	wf.Run()

	wantContent := map[string]string{"a": `A1
A2
`,
		"b": `B1
B2
B3
`,
		"c": `C1
`}

	for tag, fn := range map[string]string{
		"a": tempDir + "/concat.txt.dataset_name_a",
		"b": tempDir + "/concat.txt.dataset_name_b",
		"c": tempDir + "/concat.txt.dataset_name_c",
	} {
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			t.Fatalf("File does not exist: %s", fn)
		}
		haveContentByte, err := ioutil.ReadFile(fn)
		if err != nil {
			panic(err)
		}
		haveContent := string(haveContentByte)

		if haveContent != wantContent[tag] {
			t.Fatalf("Wanted: %s but got %s", wantContent[tag], haveContent)
		}
	}

	cleanUpTestFiles()
}

func createTestFiles() {
	os.MkdirAll(tempDir, 0744)
	for name, content := range testFiles {
		ioutil.WriteFile(tempDir+"/"+name, []byte(content), 0644)
	}
}

func cleanUpTestFiles() {
	for name := range testFiles {
		err := os.Remove(tempDir + "/" + name)
		if err != nil {
			panic(err)
		}
	}
}
