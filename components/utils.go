package components

import (
	"errors"
	"fmt"
)

func errWrap(err error, msg string) error {
	return errors.New(msg + "\nOriginal error: " + err.Error())
}

func errWrapf(err error, msg string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, v...) + "\nOriginal error: " + err.Error())
}
