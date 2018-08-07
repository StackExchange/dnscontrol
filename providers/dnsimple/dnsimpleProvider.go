package dnsimple

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/pkg/errors"

	dnsimpleapi "github.com/dnsimple/dnsimple-go/dnsimple"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("DNSimple does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
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

// DnsimpleApi is the handle for this provider.
type DnsimpleApi struct {
	AccountToken string // The account access token
	BaseURL      string // An alternate base URI
	accountID    string // Account id cache
}

// GetNameservers returns the name servers for a domain.
func (c *DnsimpleApi) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers(defaultNameServerNames), nil
}

// GetDomainCorrections returns corrections that update a domain.
func (c *DnsimpleApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}
	dc.Punycode()
	records, err := c.getRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	var actual []*models.RecordConfig
	for _, r := range records {
		if r.Type == "SOA" || r.Type == "NS" {
			continue
		}
		if r.Name == "" {
			r.Name = "@"
		}
		if r.Type == "CNAME" || r.Type == "MX" || r.Type == "ALIAS" || r.Type == "SRV" {
			r.Content += "."
		}
		// dnsimple adds these odd txt records that mirror the alias records.
		// they seem to manage them on deletes and things, so we'll just pretend they don't exist
		if r.Type == "TXT" && strings.HasPrefix(r.Content, "ALIAS for ") {
			continue
		}
		rec := &models.RecordConfig{
			TTL:      uint32(r.TTL),
			Original: r,
		}
		rec.SetLabel(r.Name, dc.Name)
		switch rtype := r.Type; rtype {
		case "ALIAS", "URL":
			rec.Type = r.Type
			rec.SetTarget(r.Content)
		case "MX":
			if err := rec.SetTargetMX(uint16(r.Priority), r.Content); err != nil {
				panic(errors.Wrap(err, "unparsable record received from dnsimple"))
			}
		default:
			if err := rec.PopulateFromString(r.Type, r.Content, dc.Name); err != nil {
				panic(errors.Wrap(err, "unparsable record received from dnsimple"))
			}
		}
		actual = append(actual, rec)
	}
	removeOtherNS(dc)
	// dc.Filter(func(r *models.RecordConfig) bool {
	// 	if r.Type == "CAA" || r.Type == "SRV" {
	// 		r.MergeToTarget()
	// 	}
	// 	return true
	// })

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(actual)

	for _, del := range delete {
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
		new := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(&old, new, dc.Name),
		})
	}

	return corrections, nil
}

// GetRegistrarCorrections returns corrections that update a domain's registrar.
func (c *DnsimpleApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
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

// DNSimple calls

func (c *DnsimpleApi) getClient() *dnsimpleapi.Client {
	client := dnsimpleapi.NewClient(dnsimpleapi.NewOauthTokenCredentials(c.AccountToken))
	if c.BaseURL != "" {
		client.BaseURL = c.BaseURL
	}
	return client
}

func (c *DnsimpleApi) getAccountID() (string, error) {
	if c.accountID == "" {
		client := c.getClient()
		whoamiResponse, err := client.Identity.Whoami()
		if err != nil {
			return "", err
		}
		if whoamiResponse.Data.User != nil && whoamiResponse.Data.Account == nil {
			return "", errors.Errorf("DNSimple token appears to be a user token. Please supply an account token")
		}
		c.accountID = strconv.Itoa(whoamiResponse.Data.Account.ID)
	}
	return c.accountID, nil
}

func (c *DnsimpleApi) getRecords(domainName string) ([]dnsimpleapi.ZoneRecord, error) {
	client := c.getClient()

	accountID, err := c.getAccountID()
	if err != nil {
		return nil, err
	}

	opts := &dnsimpleapi.ZoneRecordListOptions{}
	recs := []dnsimpleapi.ZoneRecord{}
	opts.Page = 1
	for {
		recordsResponse, err := client.Zones.ListRecords(accountID, domainName, opts)
		if err != nil {
			return nil, err
		}
		recs = append(recs, recordsResponse.Data...)
		pg := recordsResponse.Pagination
		if pg.CurrentPage == pg.TotalPages {
			break
		}
		opts.Page++
	}

	return recs, nil
}

// Returns the name server names that should be used. If the domain is registered
// then this method will return the delegation name servers. If this domain
// is hosted only, then it will return the default DNSimple name servers.
func (c *DnsimpleApi) getNameservers(domainName string) ([]string, error) {
	client := c.getClient()

	accountID, err := c.getAccountID()
	if err != nil {
		return nil, err
	}

	domainResponse, err := client.Domains.GetDomain(accountID, domainName)
	if err != nil {
		return nil, err
	}

	if domainResponse.Data.State == stateRegistered {

		delegationResponse, err := client.Registrar.GetDomainDelegation(accountID, domainName)
		if err != nil {
			return nil, err
		}

		return *delegationResponse.Data, nil
	}
	return defaultNameServerNames, nil
}

// Returns a function that can be invoked to change the delegation of the domain to the given name server names.
func (c *DnsimpleApi) updateNameserversFunc(nameServerNames []string, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}

		nameServers := dnsimpleapi.Delegation(nameServerNames)

		_, err = client.Registrar.ChangeDomainDelegation(accountID, domainName, &nameServers)
		if err != nil {
			return err
		}

		return nil
	}
}

// Returns a function that can be invoked to create a record in a zone.
func (c *DnsimpleApi) createRecordFunc(rc *models.RecordConfig, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}
		record := dnsimpleapi.ZoneRecord{
			Name:     rc.GetLabel(),
			Type:     rc.Type,
			Content:  rc.GetTargetCombined(),
			TTL:      int(rc.TTL),
			Priority: int(rc.MxPreference),
		}
		_, err = client.Zones.CreateRecord(accountID, domainName, record)
		if err != nil {
			return err
		}

		return nil
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *DnsimpleApi) deleteRecordFunc(recordID int, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}

		_, err = client.Zones.DeleteRecord(accountID, domainName, recordID)
		if err != nil {
			return err
		}

		return nil

	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *DnsimpleApi) updateRecordFunc(old *dnsimpleapi.ZoneRecord, rc *models.RecordConfig, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountID, err := c.getAccountID()
		if err != nil {
			return err
		}

		record := dnsimpleapi.ZoneRecord{
			Name:     rc.GetLabel(),
			Type:     rc.Type,
			Content:  rc.GetTargetCombined(),
			TTL:      int(rc.TTL),
			Priority: int(rc.MxPreference),
		}

		_, err = client.Zones.UpdateRecord(accountID, domainName, old.ID, record)
		if err != nil {
			return err
		}

		return nil
	}
}

// constructors

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf, nil)
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf, metadata)
}

func newProvider(m map[string]string, metadata json.RawMessage) (*DnsimpleApi, error) {
	api := &DnsimpleApi{}
	api.AccountToken = m["token"]
	if api.AccountToken == "" {
		return nil, errors.Errorf("missing DNSimple token")
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
