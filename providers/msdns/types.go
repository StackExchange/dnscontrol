package msdns

import (
	"encoding/json"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// DNSAccessor describes a system that can access Microsoft DNS.
type DNSAccessor interface {
	Exit()
	GetDNSServerZoneAll(dnsserver string) ([]string, error)
	GetDNSZoneRecords(dnsserver, domain string) ([]nativeRecord, error)
	RecordCreate(dnsserver, domain string, rec *models.RecordConfig) error
	RecordDelete(dnsserver, domain string, rec *models.RecordConfig) error
	RecordModify(dnsserver, domain string, old, rec *models.RecordConfig) error
}

// nativeRecord the JSON received from PowerShell when listing all DNS
// records in a zone.
type nativeRecord struct {
	//CimClass              interface{} `json:"CimClass"`
	//CimInstanceProperties interface{} `json:"CimInstanceProperties"`
	//CimSystemProperties   interface{} `json:"CimSystemProperties"`
	//DistinguishedName     string      `json:"DistinguishedName"`
	//RecordClass           string      `json:"RecordClass"`
	RecordType string `json:"RecordType"`
	HostName   string `json:"HostName"`
	RecordData struct {
		CimInstanceProperties []ciProperty `json:"CimInstanceProperties"`
	} `json:"RecordData"`
	TimeToLive struct {
		TotalSeconds float64 `json:"TotalSeconds"`
	} `json:"TimeToLive"`
}

type ciProperty struct {
	Name  string          `json:"Name"`
	Value json.RawMessage `json:"Value,omitempty"`
}

type ciValueDuration struct {
	TotalSeconds float64 `json:"TotalSeconds"`
}

// NB(tlim): The above structs were created using the help of:
// Get-DnsServerResourceRecord -ZoneName example.com | where { $_.RecordType -eq "SOA" } | select $_.RecordData | ConvertTo-Json -depth 10
// and pass it to https://mholt.github.io/json-to-go/
