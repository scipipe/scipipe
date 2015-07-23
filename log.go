package scipipe

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace     *log.Logger
	Debug     *log.Logger
	Info      *log.Logger
	Warn      *log.Logger
	Error     *log.Logger
	LogExists bool
)

func InitLog(
	traceHandle io.Writer,
	debugHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE:   ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(debugHandle,
		"DEBUG:   ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO:    ",
		log.Ldate|log.Ltime)

	Warn = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime)

	Error = log.New(errorHandle,
		"ERROR:   ",
		log.Ldate|log.Ltime)

	LogExists = true
}

func InitLogDebug() {
	InitLog(ioutil.Discard, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
}

func InitLogInfo() {
	InitLog(ioutil.Discard, ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

func InitLogWarn() {
	InitLog(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stdout, os.Stderr)
}

func InitLogError() {
	InitLog(ioutil.Discard, ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
}
