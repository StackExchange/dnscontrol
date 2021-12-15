package autodns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers/bind"
)

type ResourceRecord struct {

	// The name of the record.
	// Required: true
	Name string `json:"name"`

	// Preference of the record, need for some record types e.g. MX
	// Maximum: 65535
	Pref int32 `json:"pref,omitempty"`

	// The bind notation of the record. Only used by the zone stream task!
	Raw string `json:"raw,omitempty"`

	// TTL of the record (Optionally if not set then Default SOA TTL is used)
	TTL int64 `json:"ttl,omitempty"`

	// The type of the record, e.g. A
	// Permitted values: A, AAAA, CAA, CNAME, HINFO, MX, NAPTR, NS, PTR, SRV, TXT, ALIAS
	Type string `json:"type,omitempty"`

	// The value of the record.
	Value string `json:"value,omitempty"`
}

type MainAddressRecord struct {

	// TTL of the record (Optionally if not set then Default SOA TTL is used)
	TTL int64 `json:"ttl,omitempty"`

	// The value of the record.
	Value string `json:"address,omitempty"`
}

type Zone struct {

	Origin string `json:"origin"`

	Soa * bind.SoaDefaults `json:"soa,omitempty"`

	// List of name servers
	NameServers []*models.Nameserver `json:"nameServers,omitempty"`

	// The resource records.
	// Max Items: 10000
	// Min Items: 0
	ResourceRecords []*ResourceRecord `json:"resourceRecords,omitempty"`

	// Might be set if we fetch a zone for the first time, should be migrated to ResourceRecords
	MainRecord *MainAddressRecord `json:"main,omitempty"`

	IncludeWwwForMain bool `json:"wwwInclude"`

	// Primary NameServer, needs to be passed to the system to fetch further zone info
	SystemNameServer string `json:"virtualNameServer,omitempty"`
}

type JSONResponseDataZone struct {

	// The data for the response. The type of the objects are depending on the request and are also specified in the responseObject value of the response.
	Data []*Zone `json:"data"`
}
