package components

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"testing"

	"github.com/scipipe/scipipe"
	sp "github.com/scipipe/scipipe"
)

const (
	TEMPDIR = ".tmp"
)

func TestIPSelectorSync(t *testing.T) {
	wf := sp.NewWorkflow("testwf", 4)

	err := os.Mkdir(TEMPDIR, 0755)
	scipipe.Check(err)

	for _, fName := range []string{"nonempty1", "nonempty2", "pairedwithempty"} {
		nonEmptyFile, err := os.Create(fmt.Sprintf("%s/%s.txt", TEMPDIR, fName))
		scipipe.Check(err)
		_, err = nonEmptyFile.WriteString(fmt.Sprintf("%s\n", fName))
		err = nonEmptyFile.Close()
		scipipe.Check(err)
	}

	EmptyFile, err := os.Create(TEMPDIR + "/empty.txt")
	scipipe.Check(err)
	err = EmptyFile.Close()
	scipipe.Check(err)

	fileSource1 := NewFileSource(wf, "file-source-1", TEMPDIR+"/nonempty1.txt", TEMPDIR+"/empty.txt")

	fileSource2 := NewFileSource(wf, "file-source-2", TEMPDIR+"/nonempty2.txt", TEMPDIR+"/pairedwithempty.txt")

	// This function should return true for any files to be INCLUDED and false
	// for any files that should be DROPPED
	includeNonEmptyFunc := func(ip *sp.FileIP) bool {
		return ip.Size() > 0
	}
	filterEmpty := NewIPSelectorSync(wf, "filter-empty", includeNonEmptyFunc)

	// Note that the in-ports are created on-demand, so you just access them
	// and use them
	filterEmpty.In("p1").From(fileSource1.Out())
	filterEmpty.In("p2").From(fileSource2.Out())

	fileCopier := wf.NewProc("file-copier", "cat {i:in} > {o:out}")
	fileCopier.SetOut("out", "{i:in}.copy.txt")

	// ... but as you see here, we have to create and use out-ports with the
	// same names as for the in-ports we created earlier ('foo' and 'bar')
	fileCopier.In("in").From(filterEmpty.Out("p1"))
	fileCopier.In("in").From(filterEmpty.Out("p2"))

	wf.Run()

	// Check that we don't have ANY of the downstream files of the empty one,
	// and the one that was sent "in pair" with that one, since one of them
	// (empty.txt) was empty.
	for _, fileName := range []string{TEMPDIR + "/empty.txt.copy.txt", TEMPDIR + "/pairedwithempty.txt.copy.txt"} {
		_, err := os.Stat(fileName)
		if err == nil {
			t.Errorf("File should not exist, but should be filtered out: %s\n", fileName)
		}
	}

	// Check that we DO have the outputs of the non empty pair
	for _, fileName := range []string{TEMPDIR + "/nonempty1.txt.copy.txt", TEMPDIR + "/nonempty2.txt.copy.txt"} {
		if _, err := os.Stat(fileName); errors.Is(err, fs.ErrNotExist) {
			t.Errorf("File should exist, and not be filtered out: %s\n", fileName)
		}
	}

	remErr := os.RemoveAll(TEMPDIR)
	scipipe.Check(remErr)
}

func TestIPSelectorSyncFailsOnInconsistentPortClosing(t *testing.T) {

	ensureFailsProgram("TestIPSelectorSyncFailsOnInconsistentPortClosing", func() {
		wf := sp.NewWorkflow("testwf", 4)

		// Create a non-empty file
		p1a := wf.NewProc("p1a", "echo foo > {o:foo}")
		p1a.SetOut("foo", ".tmp/foo.txt")

		// Create a non-empty file
		p2a := wf.NewProc("p2a", "echo bar > {o:bar}")
		p2a.SetOut("bar", ".tmp/bar.txt")

		// Create ANOTHER non-empty file only for one of the ports
		p1b := wf.NewProc("p1b", "echo foz > {o:foz}")
		p1b.SetOut("foz", ".tmp/foz.txt")

		// Just add a bogus func here
		includeAllFiles := func(ip *sp.FileIP) bool {
			return true
		}
		filterEmpty := NewIPSelectorSync(wf, "include-all", includeAllFiles)

		filterEmpty.In("foo").From(p1a.Out("foo"))
		filterEmpty.In("bar").From(p2a.Out("bar"))

		filterEmpty.In("foo").From(p1b.Out("foz"))

		filePadder := wf.NewProc("filepadder", "cat {i:in} > {o:out}")
		filePadder.SetOut("out", "{i:in}.padded.txt")

		filePadder.In("in").From(filterEmpty.Out("foo"))
		filePadder.In("in").From(filterEmpty.Out("bar"))

		wf.Run()
	}, t)

	cleanFiles([]string{
		".tmp/foo.txt",
		".tmp/bar.txt",
		".tmp/foz.txt",
		".tmp/foo.txt.padded.txt",
		".tmp/bar.txt.padded.txt",
		".tmp/foz.txt.padded.txt",
		"_scipipe_tmp.filepadder.3a03e1a188b1c04e5a10cc9e0d24d7898e3a87dd",
		"_scipipe_tmp.filepadder.60ff69655a7244b857982477572e5955b3e09d03",
		"_scipipe_tmp.p1a.da39a3ee5e6b4b0d3255bfef95601890afd80709",
		"_scipipe_tmp.p2a.da39a3ee5e6b4b0d3255bfef95601890afd80709",
		"_scipipe_tmp.p2b.da39a3ee5e6b4b0d3255bfef95601890afd80709",
	}...)
}

func ensureFailsProgram(testName string, crasher func(), t *testing.T) {
	// After https://talks.golang.org/2014/testing.slide#23
	if os.Getenv("BE_CRASHER") == "1" {
		crasher()
	}
	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
