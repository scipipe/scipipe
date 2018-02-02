package scipipe

import (
	"time"
)

// AuditInfo contains structured audit/provenance logging information to go with an IP
type AuditInfo struct {
	Command    string
	Params     map[string]string
	Keys       map[string]string
	ExecTimeMS time.Duration
	Upstream   map[string]*AuditInfo
}

// NewAuditInfo returns a new AuditInfo struct
func NewAuditInfo() *AuditInfo {
	return &AuditInfo{
		Command:    "",
		Params:     make(map[string]string),
		Keys:       make(map[string]string),
		ExecTimeMS: -1,
		Upstream:   make(map[string]*AuditInfo),
	}
}
