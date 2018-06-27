package main

import (
	"testing"
)

func TestNewCmd(t *testing.T) {
	initLogsTest()

	args := []string{"new", "/tmp/testwf.go"}
	err := parseFlags(args)
	if err != nil {
		t.Error("Could not parse flags:", err.Error())
	}
}
