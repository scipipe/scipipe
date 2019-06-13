package scipipe

import (
	"os"
	"testing"
)

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
