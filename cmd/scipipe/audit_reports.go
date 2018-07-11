package main

import "github.com/scipipe/scipipe"

// AuditReport is a container for data to be parsed into an audit report, in
// HTML, TeX or other format
type auditReport struct {
	FileName   string
	ScipipeVer string
	AuditInfos []*scipipe.AuditInfo
}
