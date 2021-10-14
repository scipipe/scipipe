package scipipe

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

var initTestLogsLock sync.Mutex

// --------------------------------------------------------------------------------
// Testing Helper stuff
// --------------------------------------------------------------------------------

func initTestLogs() {
	if Warning == nil {
		//InitLogDebug()
		InitLogWarning()
	}
}

func cleanFiles(fileNames ...string) {
	Debug.Println("Starting to remove files:", fileNames)
	for _, fileName := range fileNames {
		auditFileName := fileName + ".audit.json"
		if _, err := os.Stat(fileName); err == nil {
			errRem := os.Remove(fileName)
			Check(errRem)
			Debug.Println("Successfully removed file", fileName)
		}
		if _, err := os.Stat(auditFileName); err == nil {
			errRem := os.Remove(auditFileName)
			Check(errRem)
			Debug.Println("Successfully removed audit.json file", auditFileName)
		}
	}
}

func cleanFilePatterns(filePatterns ...string) {
	for _, pattern := range filePatterns {
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, file := range matches {
				if err := os.Remove(file); err != nil {
					Failf("Could not remove file: %s\nError: %v\n", file, err)
				}
			}
		} else {
			Failf("Could not glob pattern: %s\nError: %v\n", pattern, err)
		}
	}
}

func assertIsType(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(reflect.TypeOf(expected), reflect.TypeOf(actual)) {
		t.Errorf("Types do not match! (%s) and (%s)\n", reflect.TypeOf(expected).String(), reflect.TypeOf(actual).String())
	}
}

func assertNil(t *testing.T, obj interface{}, msgs ...interface{}) {
	if obj != nil {
		t.Errorf("Object is not nil: %v. Message: %v\n", obj, msgs)
	}
}

func assertNotNil(t *testing.T, obj interface{}, msgs ...interface{}) {
	if obj == nil {
		t.Errorf("Object is nil, which it should not be: %v. Message: %v\n", obj, msgs)
	}
}

func assertEqualValues(t *testing.T, expected interface{}, actual interface{}, msgs ...interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Values are not equal (Expected: %v, Actual: %v)\n", expected, actual)
	}
}
