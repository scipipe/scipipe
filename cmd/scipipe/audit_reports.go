package main

import (
	"time"

	"github.com/scipipe/scipipe"
)

// AuditReport is a container for data to be parsed into an audit report, in
// HTML, TeX or other format
type auditReport struct {
	FileName   string
	ScipipeVer string
	RunTime    time.Duration
	AuditInfos []*scipipe.AuditInfo
	ColorDef   string
}
