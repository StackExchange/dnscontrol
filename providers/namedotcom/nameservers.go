package namedotcom

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
)

var nsRegex = regexp.MustCompile(`ns([1-4])[a-z]{3}\.name\.com`)

func (n *nameDotCom) GetNameservers(domain string) ([]*models.Nameserver, error) {
	//This is an interesting edge case. Name.com expects you to SET the nameservers to ns[1-4].name.com,
	//but it will internally set it to ns1xyz.name.com, where xyz is a uniqueish 3 letters.
	//In order to avoid endless loops, we will use the unique nameservers if present, or else the generic ones if not.
	nss, err := n.getNameserversRaw(domain)
	if err != nil {
		return nil, err
	}
	toUse := []string{"ns1.name.com", "ns2.name.com", "ns3.name.com", "ns4.name.com"}
	for _, ns := range nss {
		if matches := nsRegex.FindStringSubmatch(ns); len(matches) == 2 && len(matches[1]) == 1 {
			idx := matches[1][0] - '1' //regex ensures proper range
			toUse[idx] = matches[0]
		}
	}
	return models.StringsToNameservers(toUse), nil
}

func (n *nameDotCom) getNameserversRaw(domain string) ([]string, error) {
	result := &getDomainResult{}
	if err := n.get(n.apiGetDomain(domain), result); err != nil {
		return nil, err
	}
	if err := result.getErr(); err != nil {
		return nil, err
	}
	sort.Strings(result.Nameservers)
	return result.Nameservers, nil
}

func (n *nameDotCom) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	nss, err := n.getNameserversRaw(dc.Name)
	if err != nil {
		return nil, err
	}
	foundNameservers := strings.Join(nss, ",")
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

func (n *nameDotCom) apiGetDomain(domain string) string {
	return fmt.Sprintf("%s/domain/get/%s", n.APIUrl, domain)
}
func (n *nameDotCom) apiUpdateNS(domain string) string {
	return fmt.Sprintf("%s/domain/update_nameservers/%s", n.APIUrl, domain)
}

type getDomainResult struct {
	*apiResult
	DomainName  string   `json:"domain_name"`
	Nameservers []string `json:"nameservers"`
}

func (n *nameDotCom) updateNameservers(ns []string, domain string) func() error {
	return func() error {
		dat := struct {
			Nameservers []string `json:"nameservers"`
		}{ns}
		resp, err := n.post(n.apiUpdateNS(domain), dat)
		if err != nil {
			return err
		}
		if err = resp.getErr(); err != nil {
			return err
		}
		return nil
	}
}
