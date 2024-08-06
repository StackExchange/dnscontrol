package bunnydns

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/miekg/dns/dnsutil"
	"golang.org/x/exp/slices"
)

var fqdnTypes = []recordType{recordTypeCNAME, recordTypeMX, recordTypeNS, recordTypePTR, recordTypeSRV}

func fromRecordConfig(rc *models.RecordConfig) (*record, error) {
	r := record{
		Type:  recordTypeFromString(rc.Type),
		Name:  rc.GetLabel(),
		Value: rc.GetTargetField(),
		TTL:   rc.TTL,
	}

	// While Bunny DNS does not use trailing dots, it still accepts and even preserves them for certain record types.
	// To avoid confusion, any trailing dots are removed from the record value.
	if slices.Contains(fqdnTypes, r.Type) && strings.HasSuffix(r.Value, ".") {
		r.Value = strings.TrimSuffix(r.Value, ".")
	}

	switch r.Type {
	case recordTypeNS:
		if r.Name == "" {
			r.TTL = 0
		}
	case recordTypeSRV:
		r.Priority = rc.SrvPriority
		r.Weight = rc.SrvWeight
		r.Port = rc.SrvPort
	case recordTypeCAA:
		r.Flags = rc.CaaFlag
		r.Tag = rc.CaaTag
	case recordTypeMX:
		r.Priority = rc.MxPreference
	}

	return &r, nil
}

func toRecordConfig(domain string, r *record) (*models.RecordConfig, error) {
	rc := models.RecordConfig{
		Type:     recordTypeToString(r.Type),
		TTL:      r.TTL,
		Original: r,
	}
	rc.SetLabel(r.Name, domain)

	// Bunny DNS always operates with fully-qualified names and does not use any trailing dots.
	// If a record already contains a trailing dot, which the provider UI also accepts, the record value is left as-is.
	recordValue := r.Value
	if slices.Contains(fqdnTypes, r.Type) && !strings.HasSuffix(r.Value, ".") {
		recordValue = dnsutil.AddOrigin(r.Value+".", domain)
	}

	var err error
	switch rc.Type {
	case "CAA":
		err = rc.SetTargetCAA(r.Flags, r.Tag, recordValue)
	case "MX":
		err = rc.SetTargetMX(r.Priority, recordValue)
	case "SRV":
		err = rc.SetTargetSRV(r.Priority, r.Weight, r.Port, recordValue)
	default:
		err = rc.PopulateFromStringFunc(rc.Type, recordValue, domain, nil)
	}
	if err != nil {
		return nil, err
	}

	return &rc, nil
}

type recordType int

const (
	recordTypeA        recordType = 0
	recordTypeAAAA     recordType = 1
	recordTypeCNAME    recordType = 2
	recordTypeTXT      recordType = 3
	recordTypeMX       recordType = 4
	recordTypeRedirect recordType = 5
	recordTypeFlatten  recordType = 6
	recordTypePullZone recordType = 7
	recordTypeSRV      recordType = 8
	recordTypeCAA      recordType = 9
	recordTypePTR      recordType = 10
	recordTypeScript   recordType = 11
	recordTypeNS       recordType = 12
)

func recordTypeFromString(t string) recordType {
	switch t {
	case "A":
		return recordTypeA
	case "AAAA":
		return recordTypeAAAA
	case "CNAME":
		return recordTypeCNAME
	case "TXT":
		return recordTypeTXT
	case "MX":
		return recordTypeMX
	case "REDIRECT":
		return recordTypeRedirect
	case "FLATTEN":
		return recordTypeFlatten
	case "PULL_ZONE":
		return recordTypePullZone
	case "SRV":
		return recordTypeSRV
	case "CAA":
		return recordTypeCAA
	case "PTR":
		return recordTypePTR
	case "SCRIPT":
		return recordTypeScript
	case "NS":
		return recordTypeNS
	default:
		panic(fmt.Errorf("BUNNY_DNS: rtype %v unimplemented", t))
	}
}

func recordTypeToString(t recordType) string {
	switch t {
	case recordTypeA:
		return "A"
	case recordTypeAAAA:
		return "AAAA"
	case recordTypeCNAME:
		return "CNAME"
	case recordTypeTXT:
		return "TXT"
	case recordTypeMX:
		return "MX"
	case recordTypeRedirect:
		return "REDIRECT"
	case recordTypeFlatten:
		return "FLATTEN"
	case recordTypePullZone:
		return "PULL_ZONE"
	case recordTypeSRV:
		return "SRV"
	case recordTypeCAA:
		return "CAA"
	case recordTypePTR:
		return "PTR"
	case recordTypeScript:
		return "SCRIPT"
	case recordTypeNS:
		return "NS"
	default:
		panic(fmt.Errorf("BUNNY_DNS: native rtype %v unimplemented", t))
	}
}
