package domain

import "github.com/google/uuid"

// InfoBase holds the basic domain informations returned by the domain listing
type InfoBase struct {
	// Fqdn stands for Fully Qualified Domain Name. It is the domain name managed by gandi (<domain>.<tld>)
	Fqdn string `json:"fqdn,omitempty"`
	// DomainRecordsHref contains the API URL to retrieve all records registered for the domain
	DomainRecordsHref string `json:"domain_records_href,omitempty"`
	// DomainHref contains the API URL to retrieve full DomainInfo for this domain
	DomainHref string `json:"domain_href,omitempty"`
}

// InfoExtra holds the extra domain informations returned by domain details
type InfoExtra struct {
	// ZomeUUID is the id of the zone currently configured on this domain
	ZoneUUID *uuid.UUID `json:"zone_uuid,omitempty"`
	// DomainKeysHref contains the API URL to list DNSSEC keys for this domain
	// note: DNSSEC is currently not supported by this library.
	DomainKeysHref string `json:"domain_keys_href,omitempty"`
	// ZoneHref contains the API URL to retrieve informations about the zone
	ZoneHref string `json:"zone_href,omitempty"`
	// ZoneRecordsHref contains the API URL to retrieve all records registered for the zone linked to this domain
	ZoneRecordsHref string `json:"zone_records_href,omitempty"`
}

// Info holds all domain information
type Info struct {
	*InfoBase
	*InfoExtra
}
