package scipipe

import (
	"time"
)

// AuditInfo contains structured audit/provenance logging information for a
// particular task (invocation), to go with all outgoing IPs from that task
type AuditInfo struct {
	ID          string
	ProcessName string
	Command     string
	Params      map[string]string
	Keys        map[string]string
	StartTime   time.Time
	FinishTime  time.Time
	ExecTimeMS  time.Duration
	Upstream    map[string]*AuditInfo
}

// NewAuditInfo returns a new AuditInfo struct
func NewAuditInfo() *AuditInfo {
	return &AuditInfo{
		ID:          randSeqLC(20),
		ProcessName: "",
		Command:     "",
		Params:      make(map[string]string),
		Keys:        make(map[string]string),
		ExecTimeMS:  -1,
		Upstream:    make(map[string]*AuditInfo),
	}
}
