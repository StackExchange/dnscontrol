package hostingde

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/pkg/errors"
)

var (
	errZoneNotFound = errors.Errorf("zone not found")
)

type request struct {
	AuthToken      string `json:"authToken"`
	OwnerAccountID string `json:"ownerAccountId,omitempty"`
	Filter         filter `json:"filter,omitempty"`
	Limit          uint   `json:"limit,omitempty"`
	Page           uint   `json:"page,omitempty"`

	// Update Zone
	ZoneConfig      *zoneConfig `json:"zoneConfig"`
	RecordsToAdd    []*record   `json:"recordsToAdd"`
	RecordsToModify []*record   `json:"recordsToModify"`
	RecordsToDelete []*record   `json:"recordsToDelete"`

	// Create Zone
	Records []*record `json:"records"`

	// Domain
	Domain *domainConfig `json:"domain"`
}

type filter struct {
	Field    string `json:"field"`
	Value    string `json:"value"`
	Relation string `json:"relation,omitempty"`
}

type nameserver struct {
	Name string   `json:"name"`
	IPs  []net.IP `json:"ips"`
}

type domainConfig struct {
	Name                string          `json:"name"`
	Contacts            json.RawMessage `json:"contacts"`
	Nameservers         []nameserver    `json:"nameservers"`
	TransferLockEnabled bool            `json:"transferLockEnabled"`
}

type zoneConfig struct {
	ID           string `json:"id"`
	DNSSECMode   string `json:"dnsSecMode"`
	EmailAddress string `json:"emailAddress,omitempty"`
	MasterIP     string `json:"masterIp"`
	Name         string `json:"name"` // Not required per docs, but required IRL
	NameUnicode  string `json:"nameUnicode"`
	// SOAValues    struct {
	// 	Refresh     uint32 `json:"refresh"`
	// 	Retry       uint32 `json:"retry"`
	// 	Expire      uint32 `json:"expire"`
	// 	TTL         uint32 `json:"ttl"`
	// 	NegativeTTL uint32 `json:"negativeTtl"`
	// } `json:"soaValues,omitempty"`
	Type                  string   `json:"type"`
	ZoneTransferWhitelist []string `json:"zoneTransferWhitelist"`
}

type record struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      uint32 `json:"ttl"`
	Priority uint16 `json:"priority"`
}

type response struct {
	Errors   []apiError    `json:"errors"`
	Response *responseData `json:"response"`
	Status   string        `json:"status"`
}

type apiError struct {
	Code          int    `json:"code"`
	ContextObject string `json:"contextObject"`
	ContextPath   string `json:"contextPath"`
	Text          string `json:"text"`
	Value         string `json:"value"`
}

type responseData struct {
	Data json.RawMessage `json:"data"`
	Type string          `json:"type"`

	Limit      uint `json:"limit"`
	Page       uint `json:"page"`
	TotalPages uint `json:"totalPages"`
}

func (r *record) nativeToRecord(domain string) *models.RecordConfig {
	// normalize cname,mx,ns records with dots to be consistent with our config format.
	if r.Type == "ALIAS" || r.Type == "CNAME" || r.Type == "MX" || r.Type == "NS" || r.Type == "SRV" {
		if r.Content != "." {
			r.Content = r.Content + "."
		}
	}

	rc := &models.RecordConfig{
		Type:         "",
		TTL:          r.TTL,
		MxPreference: r.Priority,
		SrvPriority:  r.Priority,
		Original:     r,
	}
	rc.SetLabelFromFQDN(r.Name, domain)

	var err error
	switch r.Type {
	case "ALIAS":
		rc.Type = r.Type
		rc.SetTarget(r.Content)
	case "NULLMX":
		err = rc.PopulateFromString("MX", "0 .", domain)
	case "MX":
		err = rc.SetTargetMX(uint16(r.Priority), r.Content)
	case "SRV":
		err = rc.SetTargetSRVPriorityString(uint16(r.Priority), r.Content)
	default:
		if err := rc.PopulateFromString(r.Type, r.Content, domain); err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}

	return rc
}

func recordToNative(rc *models.RecordConfig) *record {
	record := &record{
		Name:    rc.NameFQDN,
		Type:    rc.Type,
		Content: strings.TrimSuffix(rc.GetTargetCombined(), "."),
		TTL:     rc.TTL,
	}

	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "ALIAS", "CAA", "CNAME", "DNSKEY", "DS", "NS", "NSEC", "NSEC3", "NSEC3PARAM", "PTR", "RRSIG", "SSHFP", "TSLA":
		// Nothing special.
	case "TXT":
		if cap(rc.TxtStrings) == 1 {
			record.Content = "\"" + rc.TxtStrings[0] + "\""
		} else if cap(rc.TxtStrings) > 1 {
			record.Content = ""
			for _, str := range rc.TxtStrings {
				record.Content = record.Content + " \"" + str + "\""
			}
			record.Content = record.Content[1:len(record.Content)]
		}
	case "MX":
		record.Priority = rc.MxPreference
		record.Content = strings.TrimSuffix(rc.GetTargetField(), ".")
		if record.Content == "" {
			record.Type = "NULLMX"
			record.Priority = 10
		}
	case "SRV":
		record.Priority = rc.SrvPriority
		record.Content = fmt.Sprintf("%d %d %s", rc.SrvWeight, rc.SrvPort, strings.TrimSuffix(rc.GetTargetField(), "."))
	default:
		log.Printf("hosting.de rtype %v unimplemented", rc.Type)
	}

	return record
}
