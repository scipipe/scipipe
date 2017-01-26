package scipipe

import (
	"time"
)

type AuditInfo struct {
	Command                   string
	Params                    map[string]string
	ExecutionTimeMilliSeconds time.Duration
	UpstreamAuditInfos        map[string]*AuditInfo
}

func NewAuditInfo() *AuditInfo {
	return &AuditInfo{
		Command:                   "",
		ExecutionTimeMilliSeconds: 0,
		Params:             make(map[string]string),
		UpstreamAuditInfos: make(map[string]*AuditInfo),
	}
}
