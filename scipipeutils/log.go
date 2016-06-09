package scipipeutils

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
	Audit     *log.Logger
	Warning   *log.Logger
	Error     *log.Logger
	LogExists bool
)

// Initiate logging
func InitLog(
	traceHandle io.Writer,
	debugHandle io.Writer,
	infoHandle io.Writer,
	auditHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE   ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(debugHandle,
		"DEBUG   ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO    ",
		log.Ldate|log.Ltime)

	// This level is the one suggested to use when running scientific workflows, to retain audit
	// information
	Audit = log.New(auditHandle,
		"AUDIT   ",
		log.Ldate|log.Ltime)

	Warning = log.New(warningHandle,
		"WARNING ",
		log.Ldate|log.Ltime)

	Error = log.New(errorHandle,
		"ERROR   ",
		log.Ldate|log.Ltime)

	LogExists = true
}

// Initiate logging with level=DEBUG
func InitLogDebug() {
	InitLog(
		ioutil.Discard,
		os.Stdout,
		os.Stdout,
		os.Stdout,
		os.Stdout,
		os.Stderr,
	)
}

// Initiate logging with level=INFO
func InitLogInfo() {
	InitLog(
		ioutil.Discard,
		ioutil.Discard,
		os.Stdout,
		os.Stdout,
		os.Stdout,
		os.Stderr,
	)
}

// Initiate logging with level=AUDIT
func InitLogAudit() {
	InitLog(
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		os.Stdout,
		os.Stdout,
		os.Stderr,
	)
}

// Initiate logging with level=WARNING
func InitLogWarning() {
	InitLog(
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		os.Stdout,
		os.Stderr,
	)
}

// Initiate logging with level=ERROR
func InitLogError() {
	InitLog(
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		os.Stderr,
	)
}
