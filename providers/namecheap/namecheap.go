package namecheap

import (
	"fmt"
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

func init() {
	providers.RegisterRegistrarType("NAMECHEAP", newReg)
}

func newReg(m map[string]string) (providers.Registrar, error) {
	api := &Namecheap{}
	api.ApiUser, api.ApiKey = m["apiuser"], m["apikey"]
	if api.ApiKey == "" || api.ApiUser == "" {
		return nil, fmt.Errorf("Namecheap apikey and apiuser must be provided.")
	}
	api.client = nc.NewClient(api.ApiUser, api.ApiKey, api.ApiUser)
	return api, nil
}

func (n *Namecheap) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	info, err := n.client.DomainGetInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	//todo: sort both
	found := strings.Join(info.DNSDetails.Nameservers, ",")
	desired := ""
	for _, d := range dc.Nameservers {
		if desired != "" {
			desired += ","
		}
		desired += d.Name
	}
	if found != desired {
		parts := strings.SplitN(dc.Name, ".", 2)
		sld, tld := parts[0], parts[1]
		return []*models.Correction{
			{Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", found, desired),
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
