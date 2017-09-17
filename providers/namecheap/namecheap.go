package namecheap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	nc "github.com/billputer/go-namecheap"
)

type Namecheap struct {
	ApiKey  string
	ApiUser string
	client  *nc.Client
}

var docNotes = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot("Requires domain registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterRegistrarType("NAMECHEAP", newReg, docNotes)
	// NOTE(tlim): If in the future the DNS Service Provider is implemented,
	// most likely it will require providers.CantUseNOPURGE.
}

func newReg(m map[string]string) (providers.Registrar, error) {
	api := &Namecheap{}
	api.ApiUser, api.ApiKey = m["apiuser"], m["apikey"]
	if api.ApiKey == "" || api.ApiUser == "" {
		return nil, fmt.Errorf("Namecheap apikey and apiuser must be provided.")
	}
	api.client = nc.NewClient(api.ApiUser, api.ApiKey, api.ApiUser)

	// if BaseURL is specified in creds, use that url
	BaseURL, ok := m["BaseURL"]
	if ok {
		api.client.BaseURL = BaseURL
	}

	return api, nil
}

func (n *Namecheap) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	info, err := n.client.DomainGetInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(info.DNSDetails.Nameservers)
	found := strings.Join(info.DNSDetails.Nameservers, ",")
	desiredNs := []string{}
	for _, d := range dc.Nameservers {
		desiredNs = append(desiredNs, d.Name)
	}
	sort.Strings(desiredNs)
	desired := strings.Join(desiredNs, ",")
	if found != desired {
		parts := strings.SplitN(dc.Name, ".", 2)
		sld, tld := parts[0], parts[1]
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", found, desired),
				F: func() error {
					_, err := n.client.DomainDNSSetCustom(sld, tld, desired)
					if err != nil {
						return err
					}
					return nil
				}},
		}, nil
	}
	return nil, nil
}
