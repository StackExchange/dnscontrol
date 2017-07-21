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
	"github.com/miekg/dns/dnsutil"

	dnsimpleapi "github.com/dnsimple/dnsimple-go/dnsimple"
)

const stateRegistered = "registered"

var defaultNameServerNames = []string{
	"ns1.dnsimple.com",
	"ns2.dnsimple.com",
	"ns3.dnsimple.com",
	"ns4.dnsimple.com",
}

type DnsimpleApi struct {
	AccountToken string // The account access token
	BaseURL      string // An alternate base URI
	accountId    string // Account id cache
}

func (c *DnsimpleApi) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers(defaultNameServerNames), nil
}

func (c *DnsimpleApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}

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
		if r.Type == "CNAME" || r.Type == "MX" {
			r.Content += "."
		}
		rec := &models.RecordConfig{
			NameFQDN:     dnsutil.AddOrigin(r.Name, dc.Name),
			Type:         r.Type,
			Target:       r.Content,
			TTL:          uint32(r.TTL),
			MxPreference: uint16(r.Priority),
			Original:     r,
		}
		actual = append(actual, rec)
	}
	removeOtherNS(dc)
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

func (c *DnsimpleApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}

	nameServers, err := c.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}

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

func (c *DnsimpleApi) getAccountId() (string, error) {
	if c.accountId == "" {
		client := c.getClient()
		whoamiResponse, err := client.Identity.Whoami()
		if err != nil {
			return "", err
		}
		if whoamiResponse.Data.User != nil && whoamiResponse.Data.Account == nil {
			return "", fmt.Errorf("DNSimple token appears to be a user token. Please supply an account token")
		}
		c.accountId = strconv.Itoa(whoamiResponse.Data.Account.ID)
	}
	return c.accountId, nil
}

func (c *DnsimpleApi) getRecords(domainName string) ([]dnsimpleapi.ZoneRecord, error) {
	client := c.getClient()

	accountId, err := c.getAccountId()
	if err != nil {
		return nil, err
	}

	recordsResponse, err := client.Zones.ListRecords(accountId, domainName, nil)
	if err != nil {
		return nil, err
	}

	return recordsResponse.Data, nil
}

// Returns the name server names that should be used. If the domain is registered
// then this method will return the delegation name servers. If this domain
// is hosted only, then it will return the default DNSimple name servers.
func (c *DnsimpleApi) getNameservers(domainName string) ([]string, error) {
	client := c.getClient()

	accountId, err := c.getAccountId()
	if err != nil {
		return nil, err
	}

	domainResponse, err := client.Domains.GetDomain(accountId, domainName)
	if err != nil {
		return nil, err
	}

	if domainResponse.Data.State == stateRegistered {

		delegationResponse, err := client.Registrar.GetDomainDelegation(accountId, domainName)
		if err != nil {
			return nil, err
		}

		return *delegationResponse.Data, nil
	} else {
		return defaultNameServerNames, nil
	}
}

// Returns a function that can be invoked to change the delegation of the domain to the given name server names.
func (c *DnsimpleApi) updateNameserversFunc(nameServerNames []string, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountId, err := c.getAccountId()
		if err != nil {
			return err
		}

		nameServers := dnsimpleapi.Delegation(nameServerNames)

		_, err = client.Registrar.ChangeDomainDelegation(accountId, domainName, &nameServers)
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

		accountId, err := c.getAccountId()
		if err != nil {
			return err
		}

		record := dnsimpleapi.ZoneRecord{
			Name:     dnsutil.TrimDomainName(rc.NameFQDN, domainName),
			Type:     rc.Type,
			Content:  rc.Target,
			TTL:      int(rc.TTL),
			Priority: int(rc.MxPreference),
		}

		_, err = client.Zones.CreateRecord(accountId, domainName, record)
		if err != nil {
			return err
		}

		return nil
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *DnsimpleApi) deleteRecordFunc(recordId int, domainName string) func() error {
	return func() error {
		client := c.getClient()

		accountId, err := c.getAccountId()
		if err != nil {
			return err
		}

		_, err = client.Zones.DeleteRecord(accountId, domainName, recordId)
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

		accountId, err := c.getAccountId()
		if err != nil {
			return err
		}

		record := dnsimpleapi.ZoneRecord{
			Name:     dnsutil.TrimDomainName(rc.NameFQDN, domainName),
			Type:     rc.Type,
			Content:  rc.Target,
			TTL:      int(rc.TTL),
			Priority: int(rc.MxPreference),
		}

		_, err = client.Zones.UpdateRecord(accountId, domainName, old.ID, record)
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
		return nil, fmt.Errorf("DNSimple token must be provided.")
	}

	if m["baseurl"] != "" {
		api.BaseURL = m["baseurl"]
	}

	return api, nil
}

func init() {
	providers.RegisterRegistrarType("DNSIMPLE", newReg)
	providers.RegisterDomainServiceProviderType("DNSIMPLE", newDsp, providers.CanUsePTR)
}

// remove all non-dnsimple NS records from our desired state.
// if any are found, print a warning
func removeOtherNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" {
			// apex NS inside dnsimple are expected.
			if rec.NameFQDN == dc.Name && strings.HasSuffix(rec.Target, ".dnsimple.com.") {
				continue
			}
			fmt.Printf("Warning: dnsimple.com does not allow NS records to be modified. %s will not be added.\n", rec.Target)
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
