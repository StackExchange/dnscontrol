package dnsimple

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dnsimpleapi "github.com/dnsimple/dnsimple-go/dnsimple"
	"golang.org/x/oauth2"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("DNSimple does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

func init() {
	providers.RegisterRegistrarType("DNSIMPLE", newReg)
	providers.RegisterDomainServiceProviderType("DNSIMPLE", newDsp, features)
}

const stateRegistered = "registered"

var defaultNameServerNames = []string{
	"ns1.dnsimple.com",
	"ns2.dnsimple.com",
	"ns3.dnsimple.com",
	"ns4.dnsimple.com",
}

// dnsimpleProvider is the handle for this provider.
type dnsimpleProvider struct {
	AccountToken string // The account access token
	BaseURL      string // An alternate base URI
	accountID    string // Account id cache
}

// GetNameservers returns the name servers for a domain.
func (c *dnsimpleProvider) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameServerNames)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *dnsimpleProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := c.getRecords(domain)
	if err != nil {
		return nil, err
	}

	var cleanedRecords models.Records
	for _, r := range records {
		if r.Type == "SOA" {
			continue
		}
		if r.Name == "" {
			r.Name = "@"
		}
		if r.Type == "CNAME" || r.Type == "MX" || r.Type == "ALIAS" {
			r.Content += "."
		}
		// DNSimple adds TXT records that mirror the alias records.
		// They manage them on ALIAS updates, so pretend they don't exist
		if r.Type == "TXT" && strings.HasPrefix(r.Content, "ALIAS for ") {
			continue
		}
		rec := &models.RecordConfig{
			TTL:      uint32(r.TTL),
			Original: r,
		}
		rec.SetLabel(r.Name, domain)
		switch rtype := r.Type; rtype {
		case "DNSKEY", "CDNSKEY", "CDS":
			continue
		case "ALIAS", "URL":
			rec.Type = r.Type
			if err := rec.SetTarget(r.Content); err != nil {
				return nil, fmt.Errorf("unparsable record received from dnsimple: %w", err)
			}
		case "DS":
			if err := rec.SetTargetDSString(r.Content); err != nil {
				return nil, fmt.Errorf("unparsable record received from dnsimple: %w", err)
			}
		case "MX":
			if err := rec.SetTargetMX(uint16(r.Priority), r.Content); err != nil {
				return nil, fmt.Errorf("unparsable record received from dnsimple: %w", err)
			}
		case "SRV":
			parts := strings.Fields(r.Content)
			if len(parts) == 3 {
				r.Content += "."
			}
			if err := rec.SetTargetSRVPriorityString(uint16(r.Priority), r.Content); err != nil {
				return nil, fmt.Errorf("unparsable record received from dnsimple: %w", err)
			}
		default:
			if err := rec.PopulateFromString(r.Type, r.Content, domain); err != nil {
				return nil, fmt.Errorf("unparsable record received from dnsimple: %w", err)
			}
		}
		cleanedRecords = append(cleanedRecords, rec)
	}

	return cleanedRecords, nil
}

// GetDomainCorrections returns corrections that update a domain.
func (c *dnsimpleProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}
	err := dc.Punycode()
	if err != nil {
		return nil, err
	}

	dnssecFixes, err := c.getDNSSECCorrections(dc)
	if err != nil {
		return nil, err
	}
	corrections = append(corrections, dnssecFixes...)

	records, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	actual := removeNS(records)
	removeOtherNS(dc)

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, del, modify, err := differ.IncrementalDiff(actual)
	if err != nil {
		return nil, err
	}

	for _, del := range del {
		rec := del.Existing.Original.(dnsimpleapi.ZoneRecord)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   c.deleteRecordFunc(rec.ID, dc.Name),
		})
	}

	for _, cre := range create {
		rec := cre.Desired
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   c.createRecordFunc(rec, dc.Name),
		})
	}

	for _, mod := range modify {
		old := mod.Existing.Original.(dnsimpleapi.ZoneRecord)
		rec := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(&old, rec, dc.Name),
		})
	}

	return corrections, nil
}

func removeNS(records models.Records) models.Records {
	var noNameServers models.Records
	for _, r := range records {
		if r.Type != "NS" {
			noNameServers = append(noNameServers, r)
		}
	}
	return noNameServers
}

// GetRegistrarCorrections returns corrections that update a domain's registrar.
func (c *dnsimpleProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}

	nameServers, err := c.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(nameServers)

	actual := strings.Join(nameServers, ",")

	expectedSet := []string{}
	for _, ns := range dc.Nameservers {
		expectedSet = append(expectedSet, ns.Name)
	}
	sort.Strings(expectedSet)
	expected := strings.Join(expectedSet, ",")

	if actual != expected {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", actual, expected),
				F:   c.updateNameserversFunc(expectedSet, dc.Name),
			},
		}, nil
	}

	return corrections, nil
}

// getDNSSECCorrections returns corrections that update a domain's DNSSEC state.
func (c *dnsimpleProvider) getDNSSECCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	enabled, err := c.getDnssec(dc.Name)
	if err != nil {
		return nil, err
	}

	if enabled && dc.AutoDNSSEC == "off" {
		return []*models.Correction{
			{
				Msg: "Disable DNSSEC",
				F:   func() error { _, err := c.disableDnssec(dc.Name); return err },
			},
		}, nil
	}

	if !enabled && dc.AutoDNSSEC == "on" {
		return []*models.Correction{
			{
				Msg: "Enable DNSSEC",
				F:   func() error { _, err := c.enableDnssec(dc.Name); return err },
			},
		}, nil
	}

	return []*models.Correction{}, nil
}

// DNSimple calls

func (c *dnsimpleProvider) getClient() *dnsimpleapi.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.AccountToken})
	tc := oauth2.NewClient(context.Background(), ts)

	// new client
	client := dnsimpleapi.NewClient(tc)
	client.SetUserAgent("DNSControl")

	if c.BaseURL != "" {
		client.BaseURL = c.BaseURL
	}
	return client
}

func (c *dnsimpleProvider) getAccountID() (string, error) {
	if c.accountID == "" {
		client := c.getClient()
		whoamiResponse, err := client.Identity.Whoami(context.Background())
		if err != nil {
			return "", err
		}
		if whoamiResponse.Data.User != nil && whoamiResponse.Data.Account == nil {
			return "", fmt.Errorf("DNSimple token appears to be a user token. Please supply an account token")
		}
		c.accountID = strconv.FormatInt(whoamiResponse.Data.Account.ID, 10)
	}
	return c.accountID, nil
}

func (c *dnsimpleProvider) getRecords(domainName string) ([]dnsimpleapi.ZoneRecord, error) {
	client := c.getClient()

	accountID, err := c.getAccountID()
	if err != nil {
		return nil, err
	}

	opts := &dnsimpleapi.ZoneRecordListOptions{}
	recs := []dnsimpleapi.ZoneRecord{}
	page := 1
	for {
		opts.Page = &page
		recordsResponse, err := client.Zones.ListRecords(context.Background(), accountID, domainName, opts)
		if err != nil {
			return nil, err
		}
		recs = append(recs, recordsResponse.Data...)
		pg := recordsResponse.Pagination
		if pg.CurrentPage == pg.TotalPages {
			break
		}
		page++
	}

	return recs, nil
}

func (c *dnsimpleProvider) getDnssec(domainName string) (bool, error) {
	var (
		client    *dnsimpleapi.Client
		accountID string
		err       error
	)
	client = c.getClient()
	if accountID, err = c.getAccountID(); err != nil {
		return false, err
	}

	dnssecResponse, err := client.Domains.GetDnssec(context.Background(), accountID, domainName)
	if err != nil {
		return false, err
	}
	if dnssecResponse.Data == nil {
		return false, nil
	}
	return dnssecResponse.Data.Enabled, nil
}

func (c *dnsimpleProvider) enableDnssec(domainName string) (bool, error) {
	var (
		client    *dnsimpleapi.Client
		accountID string
		err       error
	)
	client = c.getClient()
	if accountID, err = c.getAccountID(); err != nil {
		return false, err
	}

	dnssecResponse, err := client.Domains.EnableDnssec(context.Background(), accountID, domainName)
	if err != nil {
		return false, err
	}
	if dnssecResponse.Data == nil {
		return false, nil
	}
	return dnssecResponse.Data.Enabled, nil
}

func (c *dnsimpleProvider) disableDnssec(domainName string) (bool, error) {
	var (
		client    *dnsimpleapi.Client
		accountID string
		err       error
	)
	client = c.getClient()
	if accountID, err = c.getAccountID(); err != nil {
		return false, err
	}

	dnssecResponse, err := client.Domains.DisableDnssec(context.Background(), accountID, domainName)
	if err != nil {
		return false, err
	}
	if dnssecResponse.Data == nil {
		return false, nil
	}
	return dnssecResponse.Data.Enabled, nil
}

// Returns the name server names that should be used. If the domain is registered
// then this method will return the delegation name servers. If this domain
// is hosted only, then it will return the default DNSimple name servers.
func (c *dnsimpleProvider) getNameservers(domainName string) ([]string, error) {
	client := c.getClient()

	accountID, err := c.getAccountID()
	if err != nil {
		return nil, err
	}

	domainResponse, err := client.Domains.GetDomain(context.Background(), accountID, domainName)
	if err != nil {
		return nil, err
	}

	if domainResponse.Data.State == stateRegistered {

		delegationResponse, err := client.Registrar.GetDomainDelegation(context.Background(), accountID, domainName)
		if err != nil {
			return nil, err
		}

		return *delegationResponse.Data, nil
	}
	return defaultNameServerNames, nil
}

// Returns a function that can be invoked to change the delegation of the domain to the given name server names.
func (c *dnsimpleProvider) updateNameserversFunc(nameServerNames []string, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}

		nameServers := dnsimpleapi.Delegation(nameServerNames)

		_, err = client.Registrar.ChangeDomainDelegation(context.Background(), accountID, domainName, &nameServers)
		if err != nil {
			return err
		}

		return nil
	}
}

// Returns a function that can be invoked to create a record in a zone.
func (c *dnsimpleProvider) createRecordFunc(rc *models.RecordConfig, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}
		record := dnsimpleapi.ZoneRecordAttributes{
			Name:     dnsimpleapi.String(rc.GetLabel()),
			Type:     rc.Type,
			Content:  getTargetRecordContent(rc),
			TTL:      int(rc.TTL),
			Priority: getTargetRecordPriority(rc),
		}
		_, err = client.Zones.CreateRecord(context.Background(), accountID, domainName, record)
		if err != nil {
			return err
		}

		return nil
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *dnsimpleProvider) deleteRecordFunc(recordID int64, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}

		_, err = client.Zones.DeleteRecord(context.Background(), accountID, domainName, recordID)
		if err != nil {
			return err
		}

		return nil

	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *dnsimpleProvider) updateRecordFunc(old *dnsimpleapi.ZoneRecord, rc *models.RecordConfig, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}

		record := dnsimpleapi.ZoneRecordAttributes{
			Name:     dnsimpleapi.String(rc.GetLabel()),
			Type:     rc.Type,
			Content:  getTargetRecordContent(rc),
			TTL:      int(rc.TTL),
			Priority: getTargetRecordPriority(rc),
		}

		_, err = client.Zones.UpdateRecord(context.Background(), accountID, domainName, old.ID, record)
		if err != nil {
			return err
		}

		return nil
	}
}

// ListZones returns all the zones in an account
func (c *dnsimpleProvider) ListZones() ([]string, error) {
	client := c.getClient()
	accountID, err := c.getAccountID()
	if err != nil {
		return nil, err
	}

	var zones []string
	opts := &dnsimpleapi.ZoneListOptions{}
	page := 1
	for {
		opts.Page = &page
		zonesResponse, err := client.Zones.ListZones(context.Background(), accountID, opts)
		if err != nil {
			return nil, err
		}
		for _, zone := range zonesResponse.Data {
			zones = append(zones, zone.Name)
		}
		pg := zonesResponse.Pagination
		if pg.CurrentPage == pg.TotalPages {
			break
		}
		page++
	}
	return zones, nil
}

// constructors

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf, nil)
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf, metadata)
}

func newProvider(m map[string]string, metadata json.RawMessage) (*dnsimpleProvider, error) {
	api := &dnsimpleProvider{}
	api.AccountToken = m["token"]
	if api.AccountToken == "" {
		return nil, fmt.Errorf("missing DNSimple token")
	}

	if m["baseurl"] != "" {
		api.BaseURL = m["baseurl"]
	}

	return api, nil
}

// remove all non-dnsimple NS records from our desired state.
// if any are found, print a warning
func removeOtherNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" {
			// apex NS inside dnsimple are expected.
			if rec.GetLabelFQDN() == dc.Name && strings.HasSuffix(rec.GetTargetField(), ".dnsimple.com.") {
				continue
			}
			fmt.Printf("Warning: dnsimple.com does not allow NS records to be modified. %s will not be added.\n", rec.GetTargetField())
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

// Return the correct combined content for all special record types, Target for everything else
// Using RecordConfig.GetTargetCombined returns priority in the string, which we do not allow
func getTargetRecordContent(rc *models.RecordConfig) string {
	switch rtype := rc.Type; rtype {
	case "CAA":
		return rc.GetTargetCombined()
	case "SSHFP":
		return fmt.Sprintf("%d %d %s", rc.SshfpAlgorithm, rc.SshfpFingerprint, rc.GetTargetField())
	case "DS":
		return fmt.Sprintf("%d %d %d %s", rc.DsKeyTag, rc.DsAlgorithm, rc.DsDigestType, rc.DsDigest)
	case "SRV":
		return fmt.Sprintf("%d %d %s", rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	case "TXT":
		quoted := make([]string, len(rc.TxtStrings))
		for i := range rc.TxtStrings {
			quoted[i] = quoteDNSString(rc.TxtStrings[i])
		}
		return strings.Join(quoted, " ")
	case "NAPTR":
		return fmt.Sprintf("%d %d %s %s %s %s",
			rc.NaptrOrder, rc.NaptrPreference,
			quoteDNSString(rc.NaptrFlags), quoteDNSString(rc.NaptrService),
			quoteDNSString(rc.NaptrRegexp),
			rc.GetTargetField())
	default:
		return rc.GetTargetField()
	}
}

// Return the correct priority for the record type, 0 for records without priority
func getTargetRecordPriority(rc *models.RecordConfig) int {
	switch rtype := rc.Type; rtype {
	case "MX":
		return int(rc.MxPreference)
	case "SRV":
		return int(rc.SrvPriority)
	case "NAPTR":
		// Neither order nor preference
		return 0
	default:
		return 0
	}
}

// Return a DNS string appropriately escaped for DNSimple.
// Should include the surrounding quotes.
//
// Warning: the DNSimple API is severely underdocumented in this area.
// I know that it takes multiple quoted strings just fine, and constructs the
// DNS multiple quoted items.
// I'm not 100% on the escaping, but since it's a JSON API, JSON escaping seems
// reasonable.
// I do know that DNSimple have their own checks, so anything too crazy will
// get a "400 Validation failed" HTTP response.
func quoteDNSString(unquoted string) string {
	b, err := json.Marshal(unquoted)
	if err != nil {
		panic(fmt.Errorf("unable to marshal to JSON: %q", unquoted))
	}
	return string(b)
}
