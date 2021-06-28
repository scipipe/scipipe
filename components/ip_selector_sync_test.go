package components

import (
	"os"
	"testing"

	sp "github.com/scipipe/scipipe"
)

func TestIPSelectorSync(t *testing.T) {
	wf := sp.NewWorkflow("testwf", 4)

	// Create a non-empty file
	p1a := wf.NewProc("p1a", "echo foo > {o:foo}")
	p1a.SetOut("foo", ".tmp/foo.txt")

	// Create an EMPTY file
	p2a := wf.NewProc("p2a", "touch {o:bar}")
	p2a.SetOut("bar", ".tmp/bar.txt")

	// Create a non-empty file
	p1b := wf.NewProc("p1b", "echo foz > {o:foz}")
	p1b.SetOut("foz", ".tmp/foz.txt")

	// Create a non-empty file
	p2b := wf.NewProc("p2b", "echo baz > {o:baz}")
	p2b.SetOut("baz", ".tmp/baz.txt")

	// This function should return true for any files to be INCLUDED and false
	// for any files that should be DROPPED
	filterEmptyFilesFunc := func(ip *sp.FileIP) bool {
		return ip.Size() != 0
	}
	filterEmpty := NewIPSelectorSync(wf, "filter-empty", filterEmptyFilesFunc)
	// Note that the in-ports are created on-demand, so you just access them
	// and use them
	filterEmpty.In("foo").From(p1a.Out("foo"))
	filterEmpty.In("bar").From(p2a.Out("bar"))

	filterEmpty.In("foo").From(p1b.Out("foz"))
	filterEmpty.In("bar").From(p2b.Out("baz"))

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

	// Check that we DO have the outputs of the first second ones, which are
	// both non-empty
	for _, fileName := range []string{".tmp/foz.txt.padded.txt", ".tmp/baz.txt.padded.txt"} {
		_, err := os.Stat(fileName)
		if err != nil {
			t.Errorf("File should exist, and not be filtered out: %s\n", fileName)
		}
	}

	cleanFiles([]string{
		".tmp/foo.txt",
		".tmp/bar.txt",
		".tmp/foz.txt",
		".tmp/baz.txt",
		".tmp/foo.txt.padded.txt",
		".tmp/bar.txt.padded.txt",
		".tmp/foz.txt.padded.txt",
		".tmp/baz.txt.padded.txt",
	}...)
}
