package scipipe

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	// Trace is a log handler for extremely detailed level logs. It is so far
	// sparely used in scipipe.
	Trace *log.Logger
	// Debug is a log handler for debugging level logs
	Debug *log.Logger
	// Info is a log handler for information level logs
	Info *log.Logger
	// Audit is a log handler for audit level logs
	Audit *log.Logger
	// Warning is a log handler for warning level logs
	Warning *log.Logger
	// Error is a log handler for error level logs
	Error     *log.Logger
	logExists bool
)

// InitLog initiates logging handlers
func InitLog(
	traceHandle io.Writer,
	debugHandle io.Writer,
	infoHandle io.Writer,
	auditHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	if !logExists {
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

		logExists = true
	}
}

// InitLogDebug initiates logging with level=DEBUG
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

// InitLogInfo initiates logging with level=INFO
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

// InitLogAudit initiate logging with level=AUDIT
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

// InitLogAuditToFile initiate logging with level=AUDIT, and write that to
// fileName
func InitLogAuditToFile(filePath string) {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Println("Could not create directory: " + dir + " " + err.Error())
	}

	logFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Could not create log file: " + filePath + " " + err.Error())
	}

	multiWrite := io.MultiWriter(os.Stdout, logFile)

	InitLog(
		ioutil.Discard,
		ioutil.Discard,
		ioutil.Discard,
		multiWrite,
		multiWrite,
		multiWrite,
	)
}

// InitLogWarning initiates logging with level=WARNING
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

// InitLogError initiates logging with level=ERROR
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
