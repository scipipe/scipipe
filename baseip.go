package scipipe

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

// BaseIP implements some basic functionality expected to be useful for all IP
// types, and is supposed to be embedded into other sub-types
type BaseIP struct {
	id        string
	localPath string
	doStream  bool
	lock      *sync.Mutex
	auditInfo *AuditInfo
	SubStream *InPort
}

// NewBaseIP returns a new BaseID struct, with basic info like (not really)
// globally unique IDs
func NewBaseIP() *BaseIP {
	return &BaseIP{
		id:        randSeqLC(32),
		lock:      &sync.Mutex{},
		auditInfo: NewAuditInfo(),
		SubStream: NewInPort("in_substream"),
	}
}

// Param returns the parameter named key, from the IPs audit info
func (ip *BaseIP) Param(key string) string {
	val, ok := ip.AuditInfo().Params[key]
	if !ok {
		Error.Fatalf("Could not find parameter %s in ip\n", key)
	}
	return val
}

// Key returns the key value for the key with key k from the IPs audit info
// (don't confuse this with the keys of maps in go. Keys in this case is a
// SciPipe audit info concept)
func (ip *BaseIP) Key(key string) string {
	v, ok := ip.AuditInfo().Keys[key]
	if !ok {
		Error.Fatalf("Could not find key %s in ip\n", key)
	}
	return v
}

// Keys returns the audit info's key values
func (ip *BaseIP) Keys() map[string]string {
	return ip.AuditInfo().Keys
}

// AddKey adds the key k with value v
func (ip *BaseIP) AddKey(k string, v string) {
	ai := ip.AuditInfo()
	if ai.Keys[k] != "" && ai.Keys[k] != v {
		Error.Fatalf("Can not add value %s to existing key %s with different value %s\n", v, k, ai.Keys[k])
	}
	ai.Keys[k] = v
}

// AddKeys adds a map of keys to the IPs audit info
func (ip *BaseIP) AddKeys(keys map[string]string) {
	for k, v := range keys {
		ip.AddKey(k, v)
	}
}

// AuditInfo returns the AuditInfo struct for the IP
func (ip *BaseIP) AuditInfo() *AuditInfo {
	defer ip.lock.Unlock()
	ip.lock.Lock()
	if ip.auditInfo == nil {
		ip.auditInfo = NewAuditInfo()
		auditFileData, err := ioutil.ReadFile(ip.AuditFilePath())
		if err == nil {
			unmarshalErr := json.Unmarshal(auditFileData, ip.auditInfo)
			CheckWithMsg(unmarshalErr, "Could not unmarshal audit log file content: "+ip.AuditFilePath())
		}
	}
	return ip.auditInfo
}

// SetAuditInfo sets the AuditInfo struct for the IP
func (ip *BaseIP) SetAuditInfo(ai *AuditInfo) {
	ip.lock.Lock()
	ip.auditInfo = ai
	ip.lock.Unlock()
}

// AuditFilePath returns the file path of the audit info file for the IP
func (ip *BaseIP) AuditFilePath() string {
	return ip.Path() + ".audit.json"
}

// WriteAuditLogToFile writes the audit log to its designated file
func (ip *BaseIP) WriteAuditLogToFile() {
	auditInfo := ip.AuditInfo()
	auditInfoJSON, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
	CheckWithMsg(jsonErr, "Could not marshall JSON")
	writeErr := ioutil.WriteFile(ip.AuditFilePath(), auditInfoJSON, 0644)
	CheckWithMsg(writeErr, "Could not write audit file: "+ip.Path())
}
