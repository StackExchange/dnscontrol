package tencentdns

import (
	"fmt"

	"github.com/DNSControl/dnscontrol/v4/models"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

func nativeToRecord(r *dnspod.RecordListItem, domainName string) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		TTL:      uint32(*r.TTL),
		Original: r,
	}
	rc.SetLabel(*r.Name, domainName)

	val := *r.Value
	switch *r.Type {
	case "A", "AAAA", "CNAME", "NS", "PTR", "TXT", "CAA", "SRV":
		// These are standard types, PopulateFromStringFunc handles them.
	case "MX":
		if r.MX != nil {
			val = fmt.Sprintf("%d %s", *r.MX, *r.Value)
		}
	default:
		return nil, fmt.Errorf("unsupported record type: %s", *r.Type)
	}

	rtype := *r.Type
	if rtype == "CNAME" && *r.Name == "@" {
		rtype = "ALIAS"
	}

	if err := rc.PopulateFromStringFunc(rtype, val, domainName, txtutil.ParseQuoted); err != nil {
		return nil, err
	}

	return rc, nil
}

func recordToCreateRequest(rc *models.RecordConfig) *dnspod.CreateRecordRequest {
	req := dnspod.NewCreateRecordRequest()
	req.SubDomain = commonStringPtr(rc.GetLabel())
	req.RecordType = commonStringPtr(rc.Type)
	if rc.Type == "ALIAS" {
		req.RecordType = commonStringPtr("CNAME")
	}
	req.RecordLine = commonStringPtr("默认") // Default line

	val := rc.GetTargetCombinedFunc(txtutil.EncodeQuoted)
	if rc.Type == "MX" {
		val = rc.GetTargetField()
		req.MX = commonUint64Ptr(uint64(rc.MxPreference))
	}
	req.Value = commonStringPtr(val)
	req.TTL = commonUint64Ptr(uint64(rc.TTL))

	return req
}

func recordToModifyRequest(rc *models.RecordConfig, recordId uint64) *dnspod.ModifyRecordRequest {
	req := dnspod.NewModifyRecordRequest()
	req.RecordId = commonUint64Ptr(recordId)
	req.SubDomain = commonStringPtr(rc.GetLabel())
	req.RecordType = commonStringPtr(rc.Type)
	if rc.Type == "ALIAS" {
		req.RecordType = commonStringPtr("CNAME")
	}
	req.RecordLine = commonStringPtr("默认")

	val := rc.GetTargetCombinedFunc(txtutil.EncodeQuoted)
	if rc.Type == "MX" {
		val = rc.GetTargetField()
		req.MX = commonUint64Ptr(uint64(rc.MxPreference))
	}
	req.Value = commonStringPtr(val)
	req.TTL = commonUint64Ptr(uint64(rc.TTL))

	return req
}

// Helpers to avoid importing "common" in every file if possible, or just import it.
func commonStringPtr(s string) *string {
	return &s
}

func commonUint64Ptr(u uint64) *uint64 {
	return &u
}
