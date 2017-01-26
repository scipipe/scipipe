package scipipe

type AuditInfo struct {
	Command            string
	Params             map[string]string
	UpstreamAuditInfos map[string]*AuditInfo
}

func NewAuditInfo() *AuditInfo {
	return &AuditInfo{
		Command:            "",
		Params:             make(map[string]string),
		UpstreamAuditInfos: make(map[string]*AuditInfo),
	}
}
