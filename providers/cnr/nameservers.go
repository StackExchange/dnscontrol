package cnr

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

var defaultNameservers = []*models.Nameserver{
	{Name: "ns1.rrpproxy.net"},
	{Name: "ns2.rrpproxy.net"},
	{Name: "ns3.rrpproxy.net"},
}

var nsRegex = regexp.MustCompile(`ns([1-3]{1})[0-9]+\.rrpproxy\.net`)

// GetNameservers gets the nameservers set on a domain.
func (n *CNRClient) GetNameservers(domain string) ([]*models.Nameserver, error) {
	// NOTE: This information is taken over from HX and adapted to CNR... might be wrong...
	// This is an interesting edge case. CNR expects you to SET the nameservers to ns[1-3].rrpproxy.net,
	// but it will internally set it to (ns1xyz|ns2uvw|ns3asd).rrpproxy.net, where xyz/uvw/asd is a uniqueish number.
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

func (n *CNRClient) getNameserversRaw(domain string) ([]string, error) {
	r := n.client.Request(map[string]interface{}{
		"COMMAND": "StatusDomain",
		"DOMAIN":  domain,
	})
	code := r.GetCode()
	if code != 200 {
		return nil, n.GetCNRApiError("Could not get status for domain", domain, r)
	}
	nsColumn := r.GetColumn("NAMESERVER")
	if nsColumn == nil {
		fmt.Println("No nameservers found")
		return []string{}, nil // No nameserver assigned
	}
	ns := nsColumn.GetData()
	sort.Strings(ns)
	return ns, nil
}

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (n *CNRClient) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
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

func (n *CNRClient) updateNameservers(ns []string, domain string) func() error {
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
