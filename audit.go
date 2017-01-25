package scipipe

type AuditInfo struct {
	Command string
	Params  map[string]string
}

func NewAuditInfo() *AuditInfo {
	return &AuditInfo{}
}
