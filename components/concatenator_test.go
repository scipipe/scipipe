package components

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/scipipe/scipipe"
)

var (
	tempDir = ".tmp"
)

func TestConcatenator(t *testing.T) {
	os.MkdirAll(tempDir, 0744)

	// Create example content
	testFiles := map[string]string{"a.txt": "A A A", "b.txt": "B B B", "c.txt": "C C C"}
	for name, content := range testFiles {
		ioutil.WriteFile(tempDir+"/"+name, []byte(content), 0644)
	}

	// Create workflow
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

	wantContent := `A A A
B B B
C C C
`

	if haveContent != wantContent {
		t.Fatalf("Wanted: %s but got %s", wantContent, haveContent)
	}

	// Clean up test files
	for name := range testFiles {
		err := os.Remove(tempDir + "/" + name)
		if err != nil {
			panic(err)
		}
	}
}
