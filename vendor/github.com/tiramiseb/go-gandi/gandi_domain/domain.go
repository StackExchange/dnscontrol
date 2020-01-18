package gandi_domain

import (
	"time"

	"github.com/tiramiseb/go-gandi/internal/client"
)

type Domain struct {
	client client.Gandi
}

func New(apikey string, sharingid string, debug bool) *Domain {
	client := client.New(apikey, sharingid, debug)
	client.SetEndpoint("domain/")
	return &Domain{client: *client}
}

func NewFromClient(g client.Gandi) *Domain {
	g.SetEndpoint("domain/")
	return &Domain{client: g}
}

// Contact represents a contact associated with a domain
type Contact struct {
	Country        string `json:"country"`
	Email          string `json:"email"`
	Family         string `json:"family"`
	Given          string `json:"given"`
	StreetAddr     string `json:"streetaddr"`
	ContactType    int    `json:"type"`
	BrandNumber    string `json:"brand_number,omitempty"`
	City           string `json:"city,omitempty"`
	DataObfuscated *bool  `json:"data_obfuscated,omitempty"`
	Fax            string `json:"fax,omitempty"`
	Language       string `json:"lang,omitempty"`
	MailObfuscated *bool  `json:"mail_obfuscated,omitempty"`
	Mobile         string `json:"mobile,omitempty"`
	OrgName        string `json:"orgname,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Siren          string `json:"siren,omitempty"`
	State          string `json:"state,omitempty"`
	Validation     string `json:"validation,omitempty"`
	Zip            string `json:"zip,omitempty"`
}

// DomainResponseDates represents all the dates associated with a domain
type DomainResponseDates struct {
	RegistryCreatedAt   time.Time `json:"registry_created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	AuthInfoExpiresAt   time.Time `json:"authinfo_expires_at,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	DeletesAt           time.Time `json:"deletes_at,omitempty"`
	HoldBeginsAt        time.Time `json:"hold_begins_at,omitempty"`
	HoldEndsAt          time.Time `json:"hold_ends_at,omitempty"`
	PendingDeleteEndsAt time.Time `json:"pending_delete_ends_at,omitempty"`
	RegistryEndsAt      time.Time `json:"registry_ends_at,omitempty"`
	RenewBeginsAt       time.Time `json:"renew_begins_at,omitempty"`
	RenewEndsAt         time.Time `json:"renew_ends_at,omitempty"`
}

// NameServerConfig represents the name server configuration for a domain
type NameServerConfig struct {
	Current string   `json:"current"`
	Hosts   []string `json:"hosts,omitempty"`
}

// DomainListResponse is the response object returned by listing domains
type DomainListResponse struct {
	AutoRenew   *bool               `json:"autorenew"`
	Dates       DomainResponseDates `json:"dates"`
	DomainOwner string              `json:"domain_owner"`
	FQDN        string              `json:"fqdn"`
	FQDNUnicode string              `json:"fqdn_unicode"`
	Href        string              `json:"href"`
	ID          string              `json:"id"`
	NameServer  NameServerConfig    `json:"nameserver"`
	OrgaOwner   string              `json:"orga_owner"`
	Owner       string              `json:"owner"`
	Status      []string            `json:"status"`
	TLD         string              `json:"tld"`
	SharingID   string              `json:"sharing_id,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
}

// AutoRenew is the auto renewal information for the domain
type AutoRenew struct {
	Href     string      `json:"href"`
	Dates    []time.Time `json:"dates,omitempty"`
	Duration int         `json:"duration,omitempty"`
	Enabled  *bool       `json:"enabled,omitempty"`
	OrgID    string      `json:"org_id,omitempty"`
}

// The Organisation that owns the domain
type SharingSpace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DomainResponse describes a single domain
type DomainResponse struct {
	AutoRenew    AutoRenew           `json:"autorenew"`
	CanTLDLock   *bool               `json:"can_tld_lock"`
	Contacts     DomainContacts      `json:"contacts"`
	Dates        DomainResponseDates `json:"dates"`
	FQDN         string              `json:"fqdn"`
	FQDNUnicode  string              `json:"fqdn_unicode"`
	Href         string              `json:"href"`
	Nameservers  []string            `json:"nameservers,omitempty"`
	Services     []string            `json:"services"`
	SharingSpace SharingSpace        `json:"sharing_space"`
	Status       []string            `json:"status"`
	TLD          string              `json:"tld"`
	AuthInfo     string              `json:"authinfo,omitempty"`
	ID           string              `json:"id,omitempty"`
	SharingID    string              `json:"sharing_id,omitempty"`
	Tags         []string            `json:"tags,omitempty"`
	TrusteeRoles []string            `json:"trustee_roles,omitempty"`
}

// DomainContacts is the set of contacts associated with a Domain
type DomainContacts struct {
	Admin Contact `json:"admin,omitempty"`
	Bill  Contact `json:"bill,omitempty"`
	Owner Contact `json:"owner,omitempty"`
	Tech  Contact `json:"tech,omitempty"`
}

// Nameservers represents a list of nameservers
type Nameservers struct {
	Nameservers []string `json:"nameservers,omitempty"`
}

func (g *Domain) ListDomains() (domains []DomainListResponse, err error) {
	_, err = g.client.Get("domains", nil, &domains)
	return
}

func (g *Domain) GetDomain(domain string) (domainResponse DomainResponse, err error) {
	_, err = g.client.Get("domains/"+domain, nil, &domainResponse)
	return
}

func (g *Domain) GetNameServers(domain string) (nameservers []string, err error) {
	_, err = g.client.Get("domains/"+domain+"/nameservers", nil, &nameservers)
	return
}

// UpdateNameServers sets the list of the nameservers for a domain
func (g *Domain) UpdateNameServers(domain string, ns []string) (err error) {
	_, err = g.client.Put("domains/"+domain+"/nameservers", Nameservers{ns}, nil)
	return
}

// GetContacts returns the contact objects for a domain
func (g *Domain) GetContacts(domain string) (contacts DomainContacts, err error) {
	_, err = g.client.Get("domains/"+domain+"/contacts", nil, &contacts)
	return
}

// SetContacts returns the contact objects for a domain
func (g *Domain) SetContacts(domain string, contacts DomainContacts) (err error) {
	_, err = g.client.Patch("domains/"+domain+"/contacts", contacts, nil)
	return
}
