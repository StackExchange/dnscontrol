package easyname

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

type easynameProvider struct {
	apikey   string
	apiauth  string
	signSalt string
	domains  map[string]easynameDomain
}

type easynameResponseStatus struct {
	Type    string `json:"type"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
type easynameResponse struct {
	Timestamp int64                   `json:"timestamp"`
	Signature string                  `json:"signature"`
	Status    *easynameResponseStatus `json:"status"`
}

type easynameDomainList struct {
	easynameResponse
	Domains []easynameDomain `json:"data"`
}

type easynameDomain struct {
	Id          int    `json:"id"`
	Domain      string `json:"domain"`
	NameServer1 string `json:"nameserver1"`
	NameServer2 string `json:"nameserver2"`
	NameServer3 string `json:"nameserver3"`
	NameServer4 string `json:"nameserver4"`
	NameServer5 string `json:"nameserver5"`
	NameServer6 string `json:"nameserver6"`
}

type easynameNameserveChange struct {
	easynameResponse
	Nameservers map[string]string `json:"data"`
}

func init() {
	providers.RegisterRegistrarType("EASYNAME", newEasyname)
}

func newEasyname(m map[string]string) (providers.Registrar, error) {
	api := &easynameProvider{}

	if m["email"] == "" || m["userid"] == "" || m["apikey"] == "" || m["authsalt"] == "" || m["signsalt"] == "" {
		return nil, fmt.Errorf("missing easyname email, userid, apikey, and/or salt")
	}

	api.apikey, api.signSalt = m["apikey"], m["signsalt"]
	composed := fmt.Sprintf(m["authsalt"], m["userid"], m["email"])
	api.apiauth = hashEncodeString(composed)

	return api, nil
}

func hashEncodeString(s string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(s)))
	return base64.StdEncoding.EncodeToString([]byte(hash))
}

func (c *easynameProvider) updateNameservers(nss []string, domain int) error {
	var signature string
	enc := easynameNameserveChange{Nameservers: map[string]string{}}
	for i, ns := range nss {
		enc.Nameservers[fmt.Sprintf("nameserver%d", i+1)] = ns
		signature += ns
	}

	t := time.Now().Unix()
	enc.Timestamp = t
	signature = fmt.Sprintf("%s%d", signature, t)

	insert := len(signature) / 2
	if len(signature)%2 == 1 {
		insert++
	}
	signingkey := signature[:insert] + c.signSalt + signature[insert:]
	enc.Signature = hashEncodeString(signingkey)

	body, err := json.Marshal(enc)
	if err != nil {
		return err
	}

	httpClient := http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.easyname.com/domain/%d/nameserverchange", domain), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-User-ApiKey", c.apikey)
	req.Header.Set("X-User-Authentication", c.apiauth)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	var er *easynameResponse
	bodyString, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyString, &er)

	if er.Status.Type != "success" && er.Status.Type != "pending" {
		return fmt.Errorf("unable to update easyname nameservers (%d): %s", er.Status.Code, er.Status.Message)
	}
	return nil
}

func (c *easynameProvider) getDomain(domain string) (easynameDomain, error) {
	if c.domains == nil {
		c.fetchDomainList()
	}

	d, ok := c.domains[domain]
	if !ok {
		return easynameDomain{}, fmt.Errorf("nameservers not found for %s in easyname account", domain)
	}
	return d, nil
}

func (c *easynameProvider) fetchDomainList() error {
	c.domains = map[string]easynameDomain{}
	httpClient := http.Client{}
	req, err := http.NewRequest("GET", "https://api.easyname.com/domain", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-ApiKey", c.apikey)
	req.Header.Set("X-User-Authentication", c.apiauth)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	var domains *easynameDomainList
	json.Unmarshal(bodyString, &domains)

	for _, domain := range domains.Domains {
		c.domains[domain.Domain] = domain
	}

	return nil
}

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (c *easynameProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	domain, err := c.getDomain(dc.Name)
	if err != nil {
		return nil, err
	}

	nss := []string{}
	for _, ns := range []string{domain.NameServer1, domain.NameServer2, domain.NameServer3, domain.NameServer4, domain.NameServer5, domain.NameServer6} {
		if ns != "" {
			nss = append(nss, ns)
		}
	}
	foundNameservers := strings.Join(nss, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		expected = append(expected, ns.Name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
				F: func() error {
					return c.updateNameservers(expected, domain.Id)
				},
			},
		}, nil
	}
	return nil, nil
}
