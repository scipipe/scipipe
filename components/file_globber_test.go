package components

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"io/ioutil"
	"log"

	"github.com/scipipe/scipipe"
)

func TestFileGlobber(t *testing.T) {
	letters := []string{"a", "b", "c"}

	tmpDir, err := ioutil.TempDir("", "TestFileGlobber")
	if err != nil {
		log.Fatal("could not create tmpDir: ", err)
	}
	// Create files to glob
	for _, s := range letters {
		fName := filepath.Join(tmpDir, "globfile_"+s+".txt")
		f, err := os.Create(fName)
		if err != nil {
			log.Fatalf("File could not be created: %s\n", fName)
		}
		f.WriteString(s)
	}

	// Create workflow
	wf := scipipe.NewWorkflow("wf", 4)

	globber := NewFileGlobber(wf, "globber_dependent", filepath.Join(tmpDir, "globfile_*.txt"))

	copyer := wf.NewProc("copyer", "cat {i:in} > {o:out}")
	copyer.SetOut("out", "{i:in}.copy")
	copyer.In("in").From(globber.Out())

	wf.Run()

	for _, s := range letters {
		filePath := fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.copy"), s)
		expectedSArr, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal("File did not exist: " + filePath)
		}
		expectedS := string(expectedSArr)
		if expectedS != s {
			t.Fatalf("Content of file %s was '%s' and not as expected '%s'", filePath, s, expectedS)
		}
	}

	// Clean up files
	for _, s := range letters {
		filePath := fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt"), s)
		os.Remove(filePath)
		filePath = fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.audit.json"), s)
		os.Remove(filePath)
		filePath = fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.copy"), s)
		os.Remove(filePath)
		filePath = fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.copy.audit.json"), s)
		os.Remove(filePath)
	}
}

func TestFileGlobberDependent(t *testing.T) {
	letters := []string{"a", "b", "c"}
	tmpDir, err := ioutil.TempDir("", "TestFileGlobberDependent")
	if err != nil {
		log.Fatal("could not create tmpDir: ", err)
	}
	// Create files to glob
	for _, s := range letters {
		fName := filepath.Join(tmpDir, "globfile_"+s+".txt")
		f, err := os.Create(fName)
		if err != nil {
			log.Fatalf("File could not be created: %s\n", fName)
		}
		f.WriteString(s)
	}

	// Create workflow
	wf := scipipe.NewWorkflow("wf", 4)
	flagFile := wf.NewProc("create_flag_file", "echo done > {o:doneflag}")
	flagFile.SetOut("doneflag", filepath.Join(tmpDir, "done.txt"))

	globber := NewFileGlobberDependent(wf, "globber_dependent", filepath.Join(tmpDir, "globfile_*.txt"))
	globber.InDependency().From(flagFile.Out("doneflag"))

	copyer := wf.NewProc("copyer", "cat {i:in} > {o:out}")
	copyer.SetOut("out", "{i:in}.copy")
	copyer.In("in").From(globber.Out())

	wf.Run()

	fileInfoFlagFile, err := os.Stat(filepath.Join(tmpDir, "done.txt"))
	if err != nil {
		log.Fatalf("Could not stat file: %s", filepath.Join(tmpDir, "done.txt"))
	}

	for _, s := range letters {
		filePath := fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.copy"), s)
		expectedSArr, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal("File did not exist: " + filePath)
		}
		expectedS := string(expectedSArr)
		if expectedS != s {
			t.Fatalf("Content of file %s was '%s' and not as expected '%s'", filePath, s, expectedS)
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Fatalf("Could not stat file: " + filePath)
		}
		if !fileInfo.ModTime().After(fileInfoFlagFile.ModTime()) {
			t.Errorf("Copied file was not created after flag file, which the globber depended on")
		}
	}

	// Clean up files
	for _, s := range letters {
		filePath := fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt"), s)
		os.Remove(filePath)
		filePath = fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.audit.json"), s)
		os.Remove(filePath)
		filePath = fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.copy"), s)
		os.Remove(filePath)
		filePath = fmt.Sprintf(filepath.Join(tmpDir, "globfile_%s.txt.copy.audit.json"), s)
		os.Remove(filePath)
	}
	os.Remove(filepath.Join(tmpDir, "done.txt"))
	os.Remove(filepath.Join(tmpDir, "done.txt.audit.json"))
}
