package hexonet

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

var defaultNameservers = []*models.Nameserver{
	{Name: "ns1.ispapi.net"},
	{Name: "ns2.ispapi.net"},
	{Name: "ns3.ispapi.net"},
}

var nsRegex = regexp.MustCompile(`ns([1-3]{1})[0-9]+\.ispapi\.net`)

// GetNameservers gets the nameservers set on a domain.
func (n *HXClient) GetNameservers(domain string) ([]*models.Nameserver, error) {
	// This is an interesting edge case. hexonet expects you to SET the nameservers to ns[1-3].ispapi.net,
	// but it will internally set it to (ns1xyz|ns2uvw|ns3asd).ispapi.net, where xyz/uvw/asd is a uniqueish number.
	// In order to avoid endless loops, we will use the unique nameservers if present, or else the generic ones if not.
	nss, err := n.getNameserversRaw(domain)
	if err != nil {
		return nil, err
	}
	toUse := []string{
		defaultNameservers[0].Name,
		defaultNameservers[1].Name,
		defaultNameservers[2].Name,
	}
	for _, ns := range nss {
		if matches := nsRegex.FindStringSubmatch(ns); len(matches) == 2 && len(matches[1]) == 1 {
			idx := matches[1][0] - '1' // regex ensures proper range
			toUse[idx] = matches[0]
		}
	}
	return models.ToNameservers(toUse)
}

func (n *HXClient) getNameserversRaw(domain string) ([]string, error) {
	r := n.client.Request(map[string]interface{}{
		"COMMAND": "StatusDomain",
		"DOMAIN":  domain,
	})
	code := r.GetCode()
	if code != 200 {
		return nil, n.GetHXApiError("Could not get status for domain", domain, r)
	}
	nsColumn := r.GetColumn("NAMESERVER")
	if nsColumn == nil {
		return nil, fmt.Errorf("error getting NAMESERVER column for domain: %s", domain)
	}
	ns := nsColumn.GetData()
	sort.Strings(ns)
	return ns, nil
}

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (n *HXClient) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
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

func (n *HXClient) updateNameservers(ns []string, domain string) func() error {
	return func() error {
		cmd := map[string]interface{}{
			"COMMAND": "ModifyDomain",
			"DOMAIN":  domain,
		}
		for idx, ns := range ns {
			cmd[fmt.Sprintf("NAMESERVER%d", idx)] = ns
		}
		response := n.client.Request(cmd)
		code := response.GetCode()
		if code != 200 {
			return fmt.Errorf("%d %s", code, response.GetDescription())
		}
		return nil
	}
}
