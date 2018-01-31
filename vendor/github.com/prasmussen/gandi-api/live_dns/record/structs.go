package record

const (
	// A is the type of record that hold a 32-bit IPv4 address, most commonly used to map hostnames to an IP address of the host
	A = "A"
	// AAAA is the type of record that hold a Returns a 128-bit IPv6 address, most commonly used to map hostnames to an IP address of the host
	AAAA = "AAAA"
	// CAA is the type of record that hold the DNS Certification Authority Authorization, constraining acceptable CAs for a host/domain
	CAA = "CAA"
	// CDS is the type of record that hold the child copy of DS record, for transfer to parent
	CDS = "CDS"
	// CNAME is the type of record that hold the alias of one name to another: the DNS lookup will continue by retrying the lookup with the new name
	CNAME = "CNAME"
	// DNAME is the type of record that hold the alias for a name and all its subnames, unlike CNAME, which is an alias for only the exact name.
	// Like a CNAME record, the DNS lookup will continue by retrying the lookup with the new name
	DNAME = "DNAME"
	// DS is the type of record that hold the record used to identify the DNSSEC signing key of a delegated zone
	DS = "DS"
	// LOC is the type of record that specifies a geographical location associated with a domain name
	LOC = "LOC"
	// MX is the type of record that maps a domain name to a list of message transfer agents for that domain
	MX = "MX"
	// NS is the type of record that delegates a DNS zone to use the given authoritative name servers
	NS = "NS"
	// PTR is the type of record that hold a pointer to a canonical name. Unlike a CNAME, DNS processing stops and just the name is returned.
	// The most common use is for implementing reverse DNS lookups, but other uses include such things as DNS-SD.
	PTR = "PTR"
	// SPF (99) (from RFC 4408) was specified as part of the Sender Policy Framework protocol as an alternative to storing SPF data in TXT records,
	// using the same format. It was later found that the majority of SPF deployments lack proper support for this record type, and support for it was discontinued in RFC 7208
	SPF = "SPF"
	// SRV is the type of record that hold the generalized service location record, used for newer protocols instead of creating protocol-specific records such as MX.
	SRV = "SRV"
	// SSHFP is the type of record that hold resource record for publishing SSH public host key fingerprints in the DNS System,
	// in order to aid in verifying the authenticity of the host. RFC 6594 defines ECC SSH keys and SHA-256 hashes
	SSHFP = "SSHFP"
	// TLSA is the type of record that hold a record for DANE.
	// record for DANE. RFC 6698 defines "The TLSA DNS resource record is used to associate a TLS server
	// certificate or public key with the domain name where the record is found, thus forming a 'TLSA certificate association'".
	TLSA = "TLSA"
	// TXT is the type of record that hold human readable text.
	// Since the early 1990s, however, this record more often carries machine-readable data,
	// such as specified by RFC 1464, opportunistic encryption, Sender Policy Framework, DKIM, DMARC, DNS-SD, etc.
	TXT = "TXT"
	// WKS is the type of record that describe well-known services supported by a host. Not used in practice.
	// The current recommendation and practice is to determine whether a service is supported on an IP address by trying to connect to it.
	// SMTP is even prohibited from using WKS records in MX processing
	WKS = "WKS"
)

// Info holds the record informations for a single record entry
type Info struct {
	// Href contains the API URL to get the record informations
	Href string `json:"rrset_href,omitempty"`
	// Name contains name of the subdomain for this record
	Name string `json:"rrset_name,omitempty"`
	// TTL contains the life time of the record.
	TTL int64 `json:"rrset_ttl,omitempty"`
	// Type contains the DNS record type
	Type string `json:"rrset_type,omitempty"`
	// Values contains the DNS values resolved by the record
	Values []string `json:"rrset_values,omitempty"`
}

// Status holds the data returned by the API in case of record creation or update
type Status struct {
	// Message is the status message returned by the gandi api
	Message string `json:"message"`
}
