package cscglobal

// Convert the provider's native record description to models.RecordConfig.

import (
	"net"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// nativeToRecordA takes an A record from DNS and returns a native RecordConfig struct.
func nativeToRecordA(nr nativeRecordA, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "A",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetIP(net.ParseIP(nr.Value).To4())
	return rc
}

// nativeToRecordCNAME takes a CNAME record from DNS and returns a native RecordConfig struct.
func nativeToRecordCNAME(nr nativeRecordCNAME, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "CNAME",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTarget(nr.Value)
	return rc
}

// nativeToRecordA takes an AAAA record from DNS and returns a native RecordConfig struct.
func nativeToRecordAAAA(nr nativeRecordAAAA, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "AAAA",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetIP(net.ParseIP(nr.Value).To16())
	return rc
}

// nativeToRecordTXT takes a TXT record from DNS and returns a native RecordConfig struct.
func nativeToRecordTXT(nr nativeRecordTXT, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "TXT",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTarget(nr.Value)
	return rc
}

// nativeToRecordMX takes a MX record from DNS and returns a native RecordConfig struct.
func nativeToRecordMX(nr nativeRecordMX, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "MX",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetMX(nr.Priority, nr.Value)
	return rc
}

// nativeToRecordNS takes a NS record from DNS and returns a native RecordConfig struct.
func nativeToRecordNS(nr nativeRecordNS, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "NS",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTarget(nr.Value)
	return rc
}

// nativeToRecordSRV takes a SRV record from DNS and returns a native RecordConfig struct.
func nativeToRecordSRV(nr nativeRecordSRV, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: "SRV",
		TTL:  nr.TTL,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetSRV(nr.Priority, nr.Weight, nr.Port, nr.Value)
	return rc
}
