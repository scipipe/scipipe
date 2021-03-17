package components

import (
	"os"
	"testing"

	sp "github.com/scipipe/scipipe"
)

func TestIPSelectorSync(t *testing.T) {
	wf := sp.NewWorkflow("testwf", 4)

	// Create a non-empty file
	p1 := wf.NewProc("p1", "echo foo > {o:foo}")
	p1.SetOut("foo", ".tmp/foo.txt")

	// Create an EMPTY file
	p2 := wf.NewProc("p2", "touch {o:bar}")
	p2.SetOut("bar", ".tmp/bar.txt")

	// This function should return true for any files to be INCLUDED and false
	// for any files that should be DROPPED
	filterEmptyFilesFunc := func(ip *sp.FileIP) bool {
		return ip.Size() != 0
	}
	filterEmpty := NewIPSelectorSync(wf, "filter-empty", filterEmptyFilesFunc)
	// Note that the in-ports are created on-demand, so you just access them
	// and use them
	filterEmpty.In("foo").From(p1.Out("foo"))
	filterEmpty.In("bar").From(p2.Out("bar"))

	filePadder := wf.NewProc("filepadder", "cat {i:in} > {o:out}")
	filePadder.SetOut("out", "{i:in}.padded.txt")
	// ... but as you see here, we have to create and use out-ports with the
	// same names as for the in-ports we created earlier ('foo' and 'bar')
	filePadder.In("in").From(filterEmpty.Out("foo"))
	filePadder.In("in").From(filterEmpty.Out("bar"))

	wf.Run()

	// Check that we don't have ANY of the downstream files of the first two
	// ones, since one of those first ones (bar.txt) was empty
	for _, fileName := range []string{".tmp/foo.txt.padded.txt", ".tmp/bar.txt.padded.txt"} {
		_, err := os.Stat(fileName)
		if err == nil {
			t.Errorf("File should not exist, but should be filtered out: %s\n", fileName)
		}
	}

	cleanFiles([]string{
		".tmp/foo.txt",
		".tmp/bar.txt",
		".tmp/foo.txt.padded.txt",
		".tmp/bar.txt.padded.txt",
	}...)
}
