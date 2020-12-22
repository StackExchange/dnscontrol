package activedir

import (
	"encoding/json"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// nativeRecord the JSON received from PowerShell
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

//type ciValueString string `json:"Value"`
//type ciValueInt int `json:"Value"`
type ciValueDuration struct {
	//Name         string  `json:"Name"`
	TotalSeconds float64 `json:"TotalSeconds"`
}

// NOTE: When creating that struct, it was helpful to view:
// Get-DnsServerResourceRecord -ZoneName example.com | where { $_.RecordType -eq "SOA" } | select $_.RecordData | ConvertTo-Json -depth 10
// and pass it to https://mholt.github.io/json-to-go/

// DNSAccessor describes a system that can access Microsoft DNS.
type DNSAccessor interface {
	Exit()
	GetDNSServerZoneAll() ([]string, error)
	GetDNSZoneRecords(domain string) ([]nativeRecord, error)
	RecordCreate(domain string, rec *models.RecordConfig) error
	RecordDelete(domain string, rec *models.RecordConfig) error
	RecordModify(domain string, old, rec *models.RecordConfig) error
}
