package main

import (
	"testing"

	"github.com/scipipe/scipipe"
)

func TestNewCmd(t *testing.T) {
	scipipe.InitLogError()
	args := []string{"new", "/tmp/testwf.go"}
	err := parseFlags(args)
	if err != nil {
		t.Error("Could not parse flags:", err.Error())
	}
}
