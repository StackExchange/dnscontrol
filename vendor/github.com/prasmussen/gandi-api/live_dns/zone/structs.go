package zone

import "github.com/google/uuid"

// Source: http://knowledgelayer.softlayer.com/faq/what-does-serial-refresh-retry-expire-minimum-and-ttl-mean

// Info holds the DNS zone informations
type Info struct {
	// Retry is the amount of time in seconds that a domain's primary name server (or servers)
	// should wait if an attempt to refresh by a secondary name server failed before
	// attempting to refresh a domain's zone with that secondary name server again.
	Retry int `json:"retry,omitempty"`
	// UUID is the zone id
	UUID *uuid.UUID `json:"uuid,omitempty"`
	// Minimum is the amount of time in seconds that a domain's resource records are valid.
	// This is also known as a minimum TTL, and can be overridden by an individual resource record's TTL
	Minimum int `json:"minimum,omitempty"`
	// Refresh is the amount of time in seconds that a secondary name server should wait to check for
	// a new copy of a DNS zone from the domain's primary name server. If a zone file has changed then
	// the secondary DNS server will update it's copy of the zone to match the primary DNS server's zone
	Refresh int `json:"refresh,omitempty"`
	// Expire is the amount of time in seconds that a secondary name server (or servers) will
	// hold a zone before it is no longer considered authoritative
	Expire int64 `json:"expire,omitempty"`
	// SharingID is currently undocumented in http://doc.livedns.gandi.net/
	// But seems to be the ID used to  https://admin.gandi.net/domain/<...>
	SharingID *uuid.UUID `json:"sharing_id,omitempty"`
	// Serial is the revision number of this zone file. Increment this number each time the zone
	// file is changed so that the changes will be distributed to any secondary DNS servers
	Serial int `json:"serial,omitempty"`
	// Email is listed but undocumented in http://doc.livedns.gandi.net/
	Email string `json:"email,omitempty"`
	// PrimaryNS is the name of the nameserver to be used for this zone
	PrimaryNS string `json:"primary_ns,omitempty"`
	// Name is the name of the zone
	Name string `json:"name,omitempty"`
	// DomainsHref contains the API URL to retrieve all domains using this zone
	DomainsHref string `json:"domains_href,omitempty"`
	// ZoneHref contains the API URL to retrieve full DomainInfo for this domain
	ZoneHref string `json:"zone_href,omitempty"`
	// ZoneRecordsHref contains the API URL to retrieve all records registered for the zone linked to this zone
	ZoneRecordsHref string `json:"zone_records_href,omitempty"`
}

// Status holds the data returned by the API in case of zone update or association to a domain
type Status struct {
	// Message is the status message returned by the gandi api
	Message string `json:"message"`
}

// CreateStatus holds the data for returned by the API zone creation
type CreateStatus struct {
	*Status
	// UUID is the created zone ID
	UUID *uuid.UUID `json:"uuid"`
}
