package netcup

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/miekg/dns/dnsutil"
	"strconv"
	"strings"
)

type request struct {
	Action string      `json:"action"`
	Param  interface{} `json:"param"`
}

type paramLogin struct {
	Key            string `json:"apikey"`
	Password       string `json:"apipassword"`
	CustomerNumber string `json:"customernumber"`
}

type paramLogout struct {
	Key            string `json:"apikey"`
	SessionId      string `json:"apisessionid"`
	CustomerNumber string `json:"customernumber"`
}

type paramGetRecords struct {
	Key            string `json:"apikey"`
	SessionId      string `json:"apisessionid"`
	CustomerNumber string `json:"customernumber"`
	DomainName     string `json:"domainname"`
}

type paramUpdateRecords struct {
	Key            string  `json:"apikey"`
	SessionId      string  `json:"apisessionid"`
	CustomerNumber string  `json:"customernumber"`
	DomainName     string  `json:"domainname"`
	RecordSet      records `json:"dnsrecordset"`
}

type records struct {
	Records []record `json:"dnsrecords"`
}

type record struct {
	Id          string `json:"id"`
	Hostname    string `json:"hostname"`
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Destination string `json:"destination"`
	Delete      bool   `json:"deleterecord"`
	State       string `json:"state"`
}

type response struct {
	ServerRequestId string          `json:"serverrequestid"`
	ClientRequestId string          `json:"clientrequestid"`
	Action          string          `json:"action"`
	Status          string          `json:"status"`
	StatusCode      int             `json:"statuscode"`
	ShortMessage    string          `json:"shortmessage"`
	LongMessage     string          `json:"longmessage"`
	Data            json.RawMessage `json:"responsedata"`
}

type responseLogin struct {
	SessionId string `json:"apisessionid"`
}

func toRecordConfig(domain string, r *record) *models.RecordConfig {
	priority, _ := strconv.ParseUint(r.Priority, 10, 32)

	rc := &models.RecordConfig{
		Type:         r.Type,
		TTL:          uint32(0),
		MxPreference: uint16(priority),
		SrvPriority:  uint16(priority),
		SrvWeight:    uint16(0),
		SrvPort:      uint16(0),
		Original:     r,
	}
	rc.SetLabel(r.Hostname, domain)

	switch rtype := r.Type; rtype { // #rtype_variations
	case "TXT":
		_ = rc.SetTargetTXT(r.Destination)
	case "NS", "SRV", "ALIAS", "CNAME", "MX":
		_ = rc.SetTarget(dnsutil.AddOrigin(r.Destination+".", domain))
	case "CAA":
		parts := strings.Split(r.Destination, " ")
		caaFlag, _ := strconv.ParseUint(parts[0], 10, 32)
		rc.CaaFlag = uint8(caaFlag)
		rc.CaaTag = parts[1]
		_ = rc.SetTarget(strings.Trim(parts[2], "\""))
	default:
		_ = rc.SetTarget(r.Destination)
	}

	return rc
}

func fromRecordConfig(in *models.RecordConfig) *record {
	rc := &record{
		Hostname:    in.GetLabel(),
		Type:        in.Type,
		Destination: in.GetTargetField(),
		Delete:      false,
		State:       "",
	}

	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "PTR", "TXT", "SOA", "ALIAS":
		// Nothing special.
	case "CNAME":
		rc.Destination = strings.TrimSuffix(in.GetTargetField(), ".")
	case "NS":
		return nil // API ignores NS records
	case "MX":
		rc.Destination = strings.TrimSuffix(in.GetTargetField(), ".")
		rc.Priority = strconv.Itoa(int(in.MxPreference))
	case "SRV":
		rc.Priority = strconv.Itoa(int(in.SrvPriority))
	case "CAA":
		rc.Destination = strconv.Itoa(int(in.CaaFlag)) + " " + in.CaaTag + " \"" + in.GetTargetField() + "\""
	case "TLSA":
		rc.Destination = strconv.Itoa(int(in.TlsaUsage)) + " " + strconv.Itoa(int(in.TlsaSelector)) + " " + strconv.Itoa(int(in.TlsaMatchingType))
	case "SSHFP":
		rc.Destination = strconv.Itoa(int(in.SshfpAlgorithm)) + " " + strconv.Itoa(int(in.SshfpFingerprint))
	default:
		msg := fmt.Sprintf("ClouDNS.toReq rtype %v unimplemented", rc.Type)
		panic(msg)
		// We panic so that we quickly find any switch statements
	}
	return rc
}
