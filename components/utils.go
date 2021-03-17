package components

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scipipe/scipipe"
)

func errWrap(err error, msg string) error {
	return errors.New(msg + "\nOriginal error: " + err.Error())
}

func errWrapf(err error, msg string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, v...) + "\nOriginal error: " + err.Error())
}

func cleanFiles(fileNames ...string) {
	for _, fileName := range fileNames {
		auditFileName := fileName + ".audit.json"
		if _, err := os.Stat(fileName); err == nil {
			os.Remove(fileName)
		}
		if _, err := os.Stat(auditFileName); err == nil {
			os.Remove(auditFileName)
		}
	}
}

func cleanFilePatterns(filePatterns ...string) {
	for _, pattern := range filePatterns {
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, file := range matches {
				if err := os.Remove(file); err != nil {
					scipipe.Failf("Could not remove file: %s\nError: %v\n", file, err)
				}
			}
		} else {
			scipipe.Failf("Could not glob pattern: %s\nError: %v\n", pattern, err)
		}
	}
}
