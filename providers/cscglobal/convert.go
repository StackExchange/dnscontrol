package cscglobal

// Convert the provider's native record description to models.RecordConfig.

import (
	"net"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// nativeToRecordA takes an A record from DNS and returns a native RecordConfig struct.
func nativeToRecordA(nr nativeRecordA, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "A",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetIP(net.ParseIP(nr.Value).To4())
	return rc
}

// nativeToRecordCNAME takes a CNAME record from DNS and returns a native RecordConfig struct.
func nativeToRecordCNAME(nr nativeRecordCNAME, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "CNAME",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTarget(nr.Value)
	return rc
}

// nativeToRecordAAAA takes an AAAA record from DNS and returns a native RecordConfig struct.
func nativeToRecordAAAA(nr nativeRecordAAAA, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "AAAA",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetIP(net.ParseIP(nr.Value).To16())
	return rc
}

// nativeToRecordTXT takes a TXT record from DNS and returns a native RecordConfig struct.
func nativeToRecordTXT(nr nativeRecordTXT, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "TXT",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetTXT(nr.Value)
	return rc
}

// nativeToRecordMX takes a MX record from DNS and returns a native RecordConfig struct.
func nativeToRecordMX(nr nativeRecordMX, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "MX",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetMX(nr.Priority, nr.Value)
	return rc
}

// nativeToRecordNS takes a NS record from DNS and returns a native RecordConfig struct.
func nativeToRecordNS(nr nativeRecordNS, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "NS",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTarget(nr.Value)
	return rc
}

// nativeToRecordSRV takes a SRV record from DNS and returns a native RecordConfig struct.
func nativeToRecordSRV(nr nativeRecordSRV, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "SRV",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetSRV(nr.Priority, nr.Weight, nr.Port, nr.Value)
	return rc
}

// nativeToRecordCAA takes a CAA record from DNS and returns a native RecordConfig struct.
func nativeToRecordCAA(nr nativeRecordCAA, origin string, defaultTTL uint32) *models.RecordConfig {
	ttl := nr.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}
	rc := &models.RecordConfig{
		Type: "CAA",
		TTL:  ttl,
	}
	rc.SetLabel(nr.Key, origin)
	rc.SetTargetCAA(nr.Flag, nr.Tag, nr.Value)
	return rc
}
