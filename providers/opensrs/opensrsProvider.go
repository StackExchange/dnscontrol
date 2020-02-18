package opensrs

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/providers"

	opensrs "github.com/philhug/opensrs-go/opensrs"
)

var docNotes = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CanGetZones:            providers.Unimplemented(),
}

func init() {
	providers.RegisterRegistrarType("OPENSRS", newReg)
}

var defaultNameServerNames = []string{
	"ns1.systemdns.com",
	"ns2.systemdns.com",
	"ns3.systemdns.com",
}

type OpenSRSApi struct {
	UserName string // reseller user name
	ApiKey   string // API Key

	BaseURL string          // An alternate base URI
	client  *opensrs.Client // Client
}

func (c *OpenSRSApi) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers(defaultNameServerNames), nil
}

func (c *OpenSRSApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
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

// OpenSRS calls

func (c *OpenSRSApi) getClient() *opensrs.Client {
	return c.client
}

// Returns the name server names that should be used. If the domain is registered
// then this method will return the delegation name servers. If this domain
// is hosted only, then it will return the default OpenSRS name servers.
func (c *OpenSRSApi) getNameservers(domainName string) ([]string, error) {
	client := c.getClient()

	status, err := client.Domains.GetDomain(domainName, "status", 1)
	if err != nil {
		return nil, err
	}

	if status.Attributes.LockState == "0" {
		dom, err := client.Domains.GetDomain(domainName, "nameservers", 1)
		if err != nil {
			return nil, err
		}
		return dom.Attributes.NameserverList.ToString(), nil
	} else {
		return nil, errors.New("Domain is locked")
	}
}

// Returns a function that can be invoked to change the delegation of the domain to the given name server names.
func (c *OpenSRSApi) updateNameserversFunc(nameServerNames []string, domainName string) func() error {
	return func() error {
		client := c.getClient()

		_, err := client.Domains.UpdateDomainNameservers(domainName, nameServerNames)
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

func newProvider(m map[string]string, metadata json.RawMessage) (*OpenSRSApi, error) {
	api := &OpenSRSApi{}
	api.ApiKey = m["apikey"]

	if api.ApiKey == "" {
		return nil, fmt.Errorf("OpenSRS apikey must be provided.")
	}

	api.UserName = m["username"]
	if api.UserName == "" {
		return nil, fmt.Errorf("OpenSRS username key must be provided.")
	}

	if m["baseurl"] != "" {
		api.BaseURL = m["baseurl"]
	}

	api.client = opensrs.NewClient(opensrs.NewApiKeyMD5Credentials(api.UserName, api.ApiKey))
	if api.BaseURL != "" {
		api.client.BaseURL = api.BaseURL
	}

	return api, nil
}
