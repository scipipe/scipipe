package scipipe

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatCommand(t *testing.T) {

	portInfos := map[string]*PortInfo{
		"foo": &PortInfo{
			portType:  "i",
			extension: "",
			doStream:  false,
			join:      false,
			joinSep:   "",
		},
		"bar": &PortInfo{
			portType:  "i",
			extension: "",
			doStream:  false,
			join:      false,
			joinSep:   "",
		},
		"baz": &PortInfo{
			portType:  "o",
			extension: "",
			doStream:  false,
			join:      false,
			joinSep:   "",
		},
	}
	inIPs := map[string]*FileIP{
		"foo": NewFileIP("data/foofile.txt"),
		"bar": NewFileIP("barfile.txt"),
	}
	outIPs := map[string]*FileIP{
		"baz": NewFileIP("data/outfile.txt"),
	}

	for _, tt := range []struct {
		cmdPat  string
		wantCmd string
	}{
		{cmdPat: "echo {i:foo}", wantCmd: "echo ../data/foofile.txt"},
		{cmdPat: "echo {i:foo} {i:bar}", wantCmd: "echo ../data/foofile.txt ../barfile.txt"},
		{cmdPat: "cat {i:foo|basename} {i:foo} > {o:baz}", wantCmd: "cat ../foofile.txt ../data/foofile.txt > data/outfile.txt"},
		{cmdPat: "cat {i:foo} > {o:baz|%.txt}", wantCmd: "cat ../data/foofile.txt > data/outfile"},
		{cmdPat: "cat {i:foo} > {o:baz|%.txt|basename}", wantCmd: "cat ../data/foofile.txt > outfile"},
		{cmdPat: "cat {i:foo} | tee {o:baz} > {o:baz|basename|%.txt}", wantCmd: "cat ../data/foofile.txt | tee data/outfile.txt > outfile"},
		{cmdPat: "cat {i:foo|s/foo/bar/} > {o:baz|%.txt}", wantCmd: "cat ../data/barfile.txt > data/outfile"},
	} {
		gotCmd := formatCommand(tt.cmdPat, portInfos, inIPs, nil, outIPs, nil, nil, "")
		if gotCmd != tt.wantCmd {
			t.Errorf("Wanted command: '%s', but got: '%s'", tt.wantCmd, gotCmd)
		}
	}
}

func TestTempDirsExist(t *testing.T) {
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{"in1": NewFileIP("infile.txt"), "in2": NewFileIP("infile2.txt")}, nil, nil, map[string]string{"p1": "p1val", "p2": "p2val"}, nil, "", nil, 4)
	err := os.Mkdir(tsk.TempDir(), 0644)
	Check(err)
	exists := tsk.tempDirsExist()
	if !exists {
		t.Errorf("TempDirsExist returned false, even though directory exists: %s\n", tsk.TempDir())
	}
	err = os.Remove(tsk.TempDir())
	Check(err)
}

func TestTempDir(t *testing.T) {
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{"in1": NewFileIP("infile.txt"), "in2": NewFileIP("infile2.txt")}, nil, nil, map[string]string{"p1": "p1val", "p2": "p2val"}, nil, "", nil, 4)

	expected := tempDirPrefix + ".test_task.aaa94846ee057056e7f2d4d3aa2236bdf353d5a1"
	actual := tsk.TempDir()
	if actual != expected {
		t.Errorf("TempDir() was %s Expected: %s", actual, expected)
	}
}

func TestTempDirNotOver255(t *testing.T) {
	longFileName := "very_long_filename_______________________________50_______________________________________________100_______________________________________________150_______________________________________________200_______________________________________________250__255_____"
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{"in1": NewFileIP(longFileName), "in2": NewFileIP("infile2.txt")}, nil, nil, map[string]string{"p1": "p1val", "p2": "p2val"}, nil, "", nil, 4)

	actual := len(tsk.TempDir())
	maxLen := 255
	if actual > 256 {
		t.Errorf("TempDir() generated too long a string: %d chars, should be max %d chars\nString was: %s", actual, maxLen, tsk.TempDir())
	}
}

func TestExtraFilesAtomize(t *testing.T) {
	// Since Atomize calls Debug, the logger needs to be non-nil
	InitLogError()
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{}, nil, nil, map[string]string{}, nil, "", nil, 4)
	// Create extra file
	tmpDir := tsk.TempDir()
	os.MkdirAll(tmpDir, 0777)
	fName := filepath.Join(tmpDir, "letterfile_a.txt")
	_, err := os.Create(fName)
	if err != nil {
		t.Fatalf("File could not be created: %s\n", fName)
	}
	tsk.atomizeIPs()
	filePath := filepath.Join(".", "letterfile_a.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File did not exist: " + filePath)
	}
}

func TestExtraFilesAtomizeAbsolute(t *testing.T) {
	// Since Atomize calls Debug, the logger needs to be non-nil
	InitLogError()

	// Create extra file
	tmpDir, err := ioutil.TempDir("", "TestExtraFilesAtomizeAbsolute")
	if err != nil {
		t.Fatal("could not create tmpDir: ", err)
	}

	absDir := filepath.Join(tmpDir, FSRootPlaceHolder, tmpDir)
	os.MkdirAll(absDir, 0777)
	fName := filepath.Join(absDir, "letterfile_a.txt")
	_, err = os.Create(fName)
	if err != nil {
		t.Fatalf("File could not be created: %s\n", fName)
	}
	AtomizeIPs(tmpDir)
	filePath := filepath.Join(tmpDir, "letterfile_a.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File did not exist: " + filePath)
	}
	// Ensure tmpDir wasn't removed by Atomize
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Atomize removed absolute directory")
	}
}
