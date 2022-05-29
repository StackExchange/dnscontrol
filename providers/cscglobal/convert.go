package cscglobal

// Convert the provider's native record description to models.RecordConfig.

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// nativeToRecord takes a DNS record from DNS and returns a native RecordConfig struct.
func nativeToRecordA(nr nativeRecordA, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "A",
	}
	//rc.SetLabel(nr.HostName, origin)
	//rc.TTL = uint32(nr.TimeToLive.TotalSeconds)
	return rc
}

// nativeToRecordMX takes a DNS record from DNS and returns a native RecordConfig struct.
func nativeToRecordMX(nr nativeRecordMX, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "MX",
	}
	//rc.SetLabel(nr.HostName, origin)
	//rc.TTL = uint32(nr.TimeToLive.TotalSeconds)
	return rc
}
