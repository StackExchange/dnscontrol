package gidinet

import "encoding/xml"

// SOAP envelope structures for Gidinet DNS API

// SOAPEnvelope represents the SOAP envelope wrapper.
type SOAPEnvelope struct {
	XMLName xml.Name  `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    *SOAPBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

// SOAPBody represents the SOAP body.
type SOAPBody struct {
	XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Content  any
	Fault    *SOAPFault `xml:"Fault,omitempty"`
	InnerXML []byte     `xml:",innerxml"`
}

// SOAPFault represents a SOAP fault.
type SOAPFault struct {
	XMLName     xml.Name `xml:"Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	Detail      string   `xml:"detail,omitempty"`
}

// DNSRecord represents a DNS record in the Gidinet API.
type DNSRecord struct {
	DomainName string `xml:"DomainName"`
	HostName   string `xml:"HostName"`
	RecordType string `xml:"RecordType"`
	Data       string `xml:"Data"`
	TTL        int    `xml:"TTL"`
	Priority   int    `xml:"Priority"`
}

// DNSRecordListItem represents a DNS record returned from recordGetList.
type DNSRecordListItem struct {
	DomainName       string `xml:"DomainName"`
	HostName         string `xml:"HostName"`
	RecordType       string `xml:"RecordType"`
	Data             string `xml:"Data"`
	TTL              int    `xml:"TTL"`
	Priority         int    `xml:"Priority"`
	ReadOnly         bool   `xml:"ReadOnly"`
	Suspended        bool   `xml:"Suspended"`
	SuspensionReason string `xml:"SuspensionReason"`
}

// --- Request structures ---

// RecordGetListRequest is the request for recordGetList.
type RecordGetListRequest struct {
	XMLName            xml.Name `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordGetList"`
	AccountUsername    string   `xml:"accountUsername"`
	AccountPasswordB64 string   `xml:"accountPasswordB64"`
	DomainName         string   `xml:"domainName"`
}

// RecordAddRequest is the request for recordAdd.
type RecordAddRequest struct {
	XMLName            xml.Name   `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordAdd"`
	AccountUsername    string     `xml:"accountUsername"`
	AccountPasswordB64 string     `xml:"accountPasswordB64"`
	Record             *DNSRecord `xml:"record"`
}

// RecordUpdateRequest is the request for recordUpdate.
type RecordUpdateRequest struct {
	XMLName            xml.Name   `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordUpdate"`
	AccountUsername    string     `xml:"accountUsername"`
	AccountPasswordB64 string     `xml:"accountPasswordB64"`
	OldRecord          *DNSRecord `xml:"oldRecord"`
	NewRecord          *DNSRecord `xml:"newRecord"`
}

// RecordDeleteRequest is the request for recordDelete.
type RecordDeleteRequest struct {
	XMLName            xml.Name   `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordDelete"`
	AccountUsername    string     `xml:"accountUsername"`
	AccountPasswordB64 string     `xml:"accountPasswordB64"`
	Record             *DNSRecord `xml:"record"`
}

// --- Response structures ---

// BaseResponse contains the common response fields.
type BaseResponse struct {
	ResultCode    int    `xml:"resultCode"`
	ResultSubCode int    `xml:"resultSubCode"`
	ResultText    string `xml:"resultText"`
}

// RecordGetListResponse is the response from recordGetList.
type RecordGetListResponse struct {
	XMLName       xml.Name             `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordGetListResponse"`
	ResultCode    int                  `xml:"recordGetListResult>resultCode"`
	ResultSubCode int                  `xml:"recordGetListResult>resultSubCode"`
	ResultText    string               `xml:"recordGetListResult>resultText"`
	ResultItems   []*DNSRecordListItem `xml:"recordGetListResult>resultItems>DNSRecordListItem"`
}

// RecordAddResponse is the response from recordAdd.
type RecordAddResponse struct {
	XMLName       xml.Name `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordAddResponse"`
	ResultCode    int      `xml:"recordAddResult>resultCode"`
	ResultSubCode int      `xml:"recordAddResult>resultSubCode"`
	ResultText    string   `xml:"recordAddResult>resultText"`
}

// RecordUpdateResponse is the response from recordUpdate.
type RecordUpdateResponse struct {
	XMLName       xml.Name `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordUpdateResponse"`
	ResultCode    int      `xml:"recordUpdateResult>resultCode"`
	ResultSubCode int      `xml:"recordUpdateResult>resultSubCode"`
	ResultText    string   `xml:"recordUpdateResult>resultText"`
}

// RecordDeleteResponse is the response from recordDelete.
type RecordDeleteResponse struct {
	XMLName       xml.Name `xml:"https://api.quickservicebox.com/DNS/DNSAPI recordDeleteResponse"`
	ResultCode    int      `xml:"recordDeleteResult>resultCode"`
	ResultSubCode int      `xml:"recordDeleteResult>resultSubCode"`
	ResultText    string   `xml:"recordDeleteResult>resultText"`
}

// Result codes from Gidinet API.
const (
	ResultCodeSuccess        = 0 // Operation succeeded
	ResultCodeAuthFailed     = 1 // Authentication failed
	ResultCodeReadOnly       = 2 // Cannot modify read-only value
	ResultCodeInvalidParams  = 3 // Invalid parameters
	ResultCodeUndefinedError = 4 // Undefined error
	ResultCodeNotFound       = 5 // Object not found
	ResultCodeInUse          = 6 // Object in use
)

// allowedTTLValues lists the TTL values supported by the Gidinet API.
var allowedTTLValues = []uint32{
	60,     // 60 seconds
	300,    // 5 minutes
	600,    // 10 minutes
	900,    // 15 minutes
	1800,   // 30 minutes
	2700,   // 45 minutes
	3600,   // 1 hour
	7200,   // 2 hours
	14400,  // 4 hours
	28800,  // 8 hours
	43200,  // 12 hours
	64800,  // 18 hours
	86400,  // 1 day
	172800, // 2 days
}

// --- CoreAPI structures for domain listing ---

// domainGetListRequest is the request for domainGetList (CoreAPI).
type domainGetListRequest struct {
	XMLName             xml.Name `xml:"http://api.quickservicebox.com/API/Beta/CoreAPI domainGetList"`
	AccountUsername     string   `xml:"accountUsername"`
	AccountPasswordB64  string   `xml:"accountPasswordB64"`
	OrderFieldID        int      `xml:"orderFieldId"`
	OrderMode           int      `xml:"orderMode"`
	PageSize            int      `xml:"pageSize"`
	PageNumber          int      `xml:"pageNumber"`
	GroupFilter         int64    `xml:"groupFilter"`
	DomainFilter        string   `xml:"domainFilter"`
	NameserversFilter   string   `xml:"nameserversFilter"`
	RegistrantContactID int64    `xml:"registrantContactID"`
	TechContactID       int64    `xml:"techContactID"`
}

// domainListItem represents a domain in the domainGetList response.
type domainListItem struct {
	DomainID            int64  `xml:"domainId"`
	DomainName          string `xml:"domainName"`
	DomainExtension     string `xml:"domainExtension"`
	ExpireDate          string `xml:"expireDate"`
	DeletionDate        string `xml:"deletionDate"`
	GroupName           string `xml:"groupName"`
	GroupID             int64  `xml:"groupId"`
	Nameservers         string `xml:"nameservers"`
	StatusCode          int    `xml:"statusCode"`
	RegistrantContactID int64  `xml:"registrantContactID"`
	TechContactID       int64  `xml:"techContactID"`
	RedirectCount       int    `xml:"redirectCount"`
	ServiceType         int    `xml:"serviceType"`
}

// domainGetListResponse is the response from domainGetList (CoreAPI).
type domainGetListResponse struct {
	XMLName           xml.Name          `xml:"http://api.quickservicebox.com/API/Beta/CoreAPI domainGetListResponse"`
	ResultCode        int               `xml:"domainGetListResult>resultCode"`
	ResultSubCode     int               `xml:"domainGetListResult>resultSubCode"`
	ResultText        string            `xml:"domainGetListResult>resultText"`
	TotalPages        int               `xml:"domainGetListResult>totalPages"`
	TotalDomains      int               `xml:"domainGetListResult>totalDomains"`
	CurrentPageNumber int               `xml:"domainGetListResult>currentPageNumber"`
	ResultItemCount   int               `xml:"domainGetListResult>resultItemCount"`
	ResultItems       []*domainListItem `xml:"domainGetListResult>resultItems>DomainListItem"`
}

// Domain status codes.
const (
	DomainStatusActive           = 0   // Active
	DomainStatusRegistering      = 1   // In registration
	DomainStatusTransferring     = 2   // In transfer
	DomainStatusExpired          = 3   // Expired
	DomainStatusRedemptionPeriod = 4   // Redemption period or deleting
	DomainStatusDeleting         = 5   // Deleting
	DomainStatusTransferringOut  = 6   // Transferring to another registrar
	DomainStatusTransferredOut   = 7   // Transferred to another registrar
	DomainStatusUndefined        = 255 // Undefined state or error
)

// --- Registrar API structures ---

// domainNameServersChangeRequest is the request for domainNameServersChange (CoreAPI).
type domainNameServersChangeRequest struct {
	XMLName              xml.Name `xml:"http://api.quickservicebox.com/API/Beta/CoreAPI domainNameServersChange"`
	AccountUsername      string   `xml:"accountUsername"`
	AccountPasswordB64   string   `xml:"accountPasswordB64"`
	Domain               string   `xml:"domain"`
	Nameservers          string   `xml:"nameservers"`          // Comma-separated, no spaces. Empty = use Gidinet default
	AdditionalParameters []string `xml:"additionalParameters"` // Not used in current version
}

// opResultItem represents an operation result item.
type opResultItem struct {
	ServiceKey        string `xml:"serviceKey"`
	ServiceHostname   string `xml:"serviceHostname"`
	ExitCode          int    `xml:"exitCode"`          // 0=completed, 1=queued, 2=failed
	AdditionalDetails string `xml:"additionalDetails"` // Error info if exitCode=2
	ResultItemID      int64  `xml:"resultItemId"`
}

// domainNameServersChangeResponse is the response from domainNameServersChange (CoreAPI).
type domainNameServersChangeResponse struct {
	XMLName       xml.Name        `xml:"http://api.quickservicebox.com/API/Beta/CoreAPI domainNameServersChangeResponse"`
	ResultCode    int             `xml:"domainNameServersChangeResult>resultCode"`
	ResultSubCode int             `xml:"domainNameServersChangeResult>resultSubCode"`
	ResultText    string          `xml:"domainNameServersChangeResult>resultText"`
	ResultItems   []*opResultItem `xml:"domainNameServersChangeResult>resultItems>OpResultItem"`
}
