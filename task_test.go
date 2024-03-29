package scipipe

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func getTestPortInfos() map[string]*PortInfo {
	return map[string]*PortInfo{
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
		"bax": &PortInfo{
			portType:  "o",
			extension: "",
			doStream:  false,
			join:      false,
			joinSep:   "",
		},
		"bay": &PortInfo{
			portType:  "o",
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
}

func getTestInIPs() map[string]*FileIP {
	fooIP, err := NewFileIP("data/foofile.txt")
	Check(err)
	barIP, err := NewFileIP("barfile.txt")
	Check(err)
	return map[string]*FileIP{
		"foo": fooIP,
		"bar": barIP,
	}
}

func getTestOutIPs() map[string]*FileIP {
	bazIP, err := NewFileIP("data/outfile.txt")
	Check(err)
	bayIP, err := NewFileIP("/tmp/scipipe/bay_outfile.txt")
	Check(err)
	baxIP, err := NewFileIP("../../ref/ref.txt")
	Check(err)
	return map[string]*FileIP{
		"bax": baxIP,
		"bay": bayIP,
		"baz": bazIP,
	}
}

func TestFormatCommand(t *testing.T) {
	portInfos := getTestPortInfos()
	inIPs := getTestInIPs()
	outIPs := getTestOutIPs()

	emptyTask := &Task{}

	for _, tt := range []struct {
		cmdPat  string
		wantCmd string
	}{
		{cmdPat: "echo {i:foo}", wantCmd: "echo ../data/foofile.txt"},
		{cmdPat: "echo {i:foo} {i:bar}", wantCmd: "echo ../data/foofile.txt ../barfile.txt"},
		{cmdPat: "cat {i:foo} > {o:baz|%.txt}", wantCmd: "cat ../data/foofile.txt > data/outfile"},
		{cmdPat: "cat {i:foo} > {o:baz|%.txt|basename}", wantCmd: "cat ../data/foofile.txt > outfile"},
		{cmdPat: "cat {i:foo} | tee {o:baz} > {o:baz|basename|%.txt}", wantCmd: "cat ../data/foofile.txt | tee data/outfile.txt > outfile"},
		{cmdPat: "cat {i:foo|s/foo/bar/} > {o:baz|%.txt}", wantCmd: "cat ../data/barfile.txt > data/outfile"},
		{cmdPat: "cat {i:foo|dirname}/newfile.txt {i:foo} > {o:baz}", wantCmd: "cat ../data/newfile.txt ../data/foofile.txt > data/outfile.txt"},
		{cmdPat: "cat {i:foo|dirname}/some_path/{i:foo|basename}", wantCmd: "cat ../data/some_path/foofile.txt"},
		{cmdPat: "cat ../{i:foo|basename} {i:foo} > {o:baz}", wantCmd: "cat ../foofile.txt ../data/foofile.txt > data/outfile.txt"},
		{cmdPat: "cat {i:foo} | tee {o:baz} > {o:baz|dirname}/hoge/{o:baz|basename|%.txt}.out.txt", wantCmd: "cat ../data/foofile.txt | tee data/outfile.txt > data/hoge/outfile.out.txt"},
		{cmdPat: "cat {i:foo} > {o:bax}", wantCmd: "cat ../data/foofile.txt > __parent____parent__ref/ref.txt"},
		{cmdPat: "cat {i:foo} > {o:bay}", wantCmd: "cat ../data/foofile.txt > __fsroot__/tmp/scipipe/bay_outfile.txt"},
		{cmdPat: "cat {i:foo} > data/{o:bay|basename|%.txt}.csv", wantCmd: "cat ../data/foofile.txt > data/bay_outfile.csv"},
	} {
		gotCmd := emptyTask.formatCommand(tt.cmdPat, portInfos, inIPs, nil, outIPs, nil, nil, "")
		if gotCmd != tt.wantCmd {
			t.Errorf("Wanted command: '%s', but got: '%s'", tt.wantCmd, gotCmd)
		}
	}
}

func TestTempDirsExist(t *testing.T) {
	inIP1, err := NewFileIP("infile.txt")
	Check(err)
	inIP2, err := NewFileIP("infile2.txt")
	Check(err)
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{"in1": inIP1, "in2": inIP2}, nil, nil, map[string]string{"p1": "p1val", "p2": "p2val"}, nil, "", nil, 4)
	err = os.Mkdir(tsk.TempDir(), 0644)
	Check(err)
	exists := tsk.tempDirsExist()
	if !exists {
		t.Errorf("TempDirsExist returned false, even though directory exists: %s\n", tsk.TempDir())
	}
	err = os.Remove(tsk.TempDir())
	Check(err)
}

func TestTempDir(t *testing.T) {
	inIP1, err := NewFileIP("infile.txt")
	Check(err)
	inIP2, err := NewFileIP("infile2.txt")
	Check(err)
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{"in1": inIP1, "in2": inIP2}, nil, nil, map[string]string{"p1": "p1val", "p2": "p2val"}, nil, "", nil, 4)

	expected := tempDirPrefix + ".test_task.59d80301296d143e0da61ed64e5756e72e268631"
	actual := tsk.TempDir()
	if actual != expected {
		t.Errorf("TempDir() was %s Expected: %s", actual, expected)
	}
}

func TestTempDirNotOver255(t *testing.T) {
	longFileName := "very_long_filename_______________________________50_______________________________________________100_______________________________________________150_______________________________________________200_______________________________________________250__255_____"
	longFileIP, err := NewFileIP(longFileName)
	inIP2, err := NewFileIP("infile2.txt")
	Check(err)
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{"in1": longFileIP, "in2": inIP2}, nil, nil, map[string]string{"p1": "p1val", "p2": "p2val"}, nil, "", nil, 4)

	actual := len(tsk.TempDir())
	maxLen := 255
	if actual > 256 {
		t.Errorf("TempDir() generated too long a string: %d chars, should be max %d chars\nString was: %s", actual, maxLen, tsk.TempDir())
	}
}

func TestExtraFilesFinalizePaths(t *testing.T) {
	// Since FinalizePaths calls Debug, the logger needs to be non-nil
	InitLogError()
	tsk := NewTask(nil, nil, "test_task", "echo foo", map[string]*FileIP{}, nil, nil, map[string]string{}, nil, "", nil, 4)
	// Create extra file
	tmpDir := tsk.TempDir()
	err := os.MkdirAll(tmpDir, 0777)
	Check(err)
	fName := filepath.Join(tmpDir, "letterfile_a.txt")
	_, errCreate := os.Create(fName)
	if errCreate != nil {
		t.Fatalf("File could not be created: %s\n", fName)
	}
	errFin := tsk.finalizePaths()
	Check(errFin)
	filePath := filepath.Join(".", "letterfile_a.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File did not exist: " + filePath)
	}
}

func TestExtraFilesFinalizePathsAbsolute(t *testing.T) {
	// Since FinalizePaths calls Debug, the logger needs to be non-nil
	InitLogError()

	// Create extra file
	tmpDir, err := ioutil.TempDir("", "TestExtraFilesFinalizePathsAbsolute")
	if err != nil {
		t.Fatal("could not create tmpDir: ", err)
	}

	absDir := filepath.Join(tmpDir, FSRootPlaceHolder, tmpDir)
	errMkdir := os.MkdirAll(absDir, 0777)
	Check(errMkdir)
	fName := filepath.Join(absDir, "letterfile_a.txt")
	_, err = os.Create(fName)
	if err != nil {
		t.Fatalf("File could not be created: %s\n", fName)
	}
	errFin := FinalizePaths(tmpDir)
	Check(errFin)
	filePath := filepath.Join(tmpDir, "letterfile_a.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File did not exist: " + filePath)
	}
	// Ensure tmpDir wasn't removed by FinalizePaths
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("FinalizePaths removed absolute directory")
	}
}

func TestReplaceParentDirsWithPlaceHolder(t *testing.T) {
	for _, tc := range []struct {
		inputStr string
		wantStr  string
	}{
		{inputStr: "../../some/dir", wantStr: "__parent____parent__some/dir"},
		{inputStr: "../", wantStr: "__parent__"},
		{inputStr: "./../", wantStr: "./__parent__"},
	} {
		haveStr := replaceParentDirsWithPlaceholder(tc.inputStr)
		if haveStr != tc.wantStr {
			t.Fatalf("String '%s' was replaced into '%s' and not '%s' as expected", tc.inputStr, haveStr, tc.wantStr)
		}
	}
}

func TestReplacePlaceholdersWithParentDirs(t *testing.T) {
	for _, tc := range []struct {
		inputStr string
		wantStr  string
	}{
		{inputStr: "__parent____parent__some/dir", wantStr: "../../some/dir"},
		{inputStr: "__parent__", wantStr: "../"},
		{inputStr: "./__parent__", wantStr: "./../"},
	} {
		haveStr := replacePlaceholdersWithParentDirs(tc.inputStr)
		if haveStr != tc.wantStr {
			t.Fatalf("String '%s' was replaced into '%s' and not '%s' as expected", tc.inputStr, haveStr, tc.wantStr)
		}
	}
}
