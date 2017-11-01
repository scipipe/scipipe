package scipipe

import (
	"time"
)

type AuditInfo struct {
	Command    string
	Params     map[string]string
	Keys       map[string]string
	ExecTimeMS time.Duration
	Upstream   map[string]*AuditInfo
}

func NewAuditInfo() *AuditInfo {
	return &AuditInfo{
		Command:    "",
		Params:     make(map[string]string),
		Keys:       make(map[string]string),
		ExecTimeMS: -1,
		Upstream:   make(map[string]*AuditInfo),
	}
}
