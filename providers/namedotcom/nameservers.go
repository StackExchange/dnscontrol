package namedotcom

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
)

func (n *nameDotCom) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	foundNameservers, err := n.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	if defaultNsRegexp.MatchString(foundNameservers) {
		foundNameservers = "ns1.name.com,ns2.name.com,ns3.name.com,ns4.name.com"
	}
	expected := []string{}
	for _, ns := range dc.Nameservers {
		name := strings.TrimRight(ns.Name, ".")
		expected = append(expected, name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
				F:   n.updateNameservers(expected, dc.Name),
			},
		}, nil
	}
	return nil, nil
}

//even if you provide them "ns1.name.com", they will set it to "ns1qrt.name.com". This will match that pattern to see if defaults are in use.
var defaultNsRegexp = regexp.MustCompile(`ns1[a-z]{0,3}\.name\.com,ns2[a-z]{0,3}\.name\.com,ns3[a-z]{0,3}\.name\.com,ns4[a-z]{0,3}\.name\.com`)

func apiGetDomain(domain string) string {
	return fmt.Sprintf("%s/domain/get/%s", apiBase, domain)
}
func apiUpdateNS(domain string) string {
	return fmt.Sprintf("%s/domain/update_nameservers/%s", apiBase, domain)
}

type getDomainResult struct {
	*apiResult
	DomainName  string   `json:"domain_name"`
	Nameservers []string `json:"nameservers"`
}

// returns comma joined list of nameservers (in alphabetical order)
func (n *nameDotCom) getNameservers(domain string) (string, error) {
	result := &getDomainResult{}
	if err := n.get(apiGetDomain(domain), result); err != nil {
		return "", err
	}
	if err := result.getErr(); err != nil {
		return "", err
	}
	sort.Strings(result.Nameservers)
	return strings.Join(result.Nameservers, ","), nil
}

func (n *nameDotCom) updateNameservers(ns []string, domain string) func() error {
	return func() error {
		dat := struct {
			Nameservers []string `json:"nameservers"`
		}{ns}
		resp, err := n.post(apiUpdateNS(domain), dat)
		if err != nil {
			return err
		}
		if err = resp.getErr(); err != nil {
			return err
		}
		return nil
	}
}
