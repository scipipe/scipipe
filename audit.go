package scipipe

import (
	"time"
)

type AuditInfo struct {
	Command    string
	Params     map[string]string
	ExecTimeMS time.Duration
	Upstream   map[string]*AuditInfo
}

func NewAuditInfo() *AuditInfo {
	return &AuditInfo{
		Command:    "",
		ExecTimeMS: 0,
		Params:     make(map[string]string),
		Upstream:   make(map[string]*AuditInfo),
	}
}
