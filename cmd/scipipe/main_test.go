package main

import (
	"os"
	"testing"
)

func TestNewCmd(t *testing.T) {
	initLogsTest()

	testWfPath := "/tmp/testwf.go"

	args := []string{"new", testWfPath}
	err := parseFlags(args)
	if err != nil {
		t.Error("Could not parse flags:", err.Error())
	}

	if _, err := os.Stat(testWfPath); os.IsNotExist(err) {
		t.Error(t, "`scipipe new` command failed to create new workflow file: "+testWfPath)
	}

	cleanFiles(t, testWfPath)
}

func cleanFiles(t *testing.T, files ...string) {
	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			t.Error(err.Error())
		}
	}
}
