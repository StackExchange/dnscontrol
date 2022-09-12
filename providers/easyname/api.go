package easyname

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type easynameResponse interface {
	GetStatus() easynameResponseStatus
}

type easynameResponseStatus struct {
	Type    string `json:"type"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type easynameResponseData struct {
	easynameResponseStatus
	Timestamp int64                   `json:"timestamp"`
	Signature string                  `json:"signature"`
	Status    *easynameResponseStatus `json:"status"`
}

func (c easynameResponseData) GetStatus() easynameResponseStatus {
	return *c.Status
}

type easynameDomainList struct {
	easynameResponseData
	Domains []easynameDomain `json:"data"`
}

type easynameDomain struct {
	ID          int    `json:"id"`
	Domain      string `json:"domain"`
	NameServer1 string `json:"nameserver1"`
	NameServer2 string `json:"nameserver2"`
	NameServer3 string `json:"nameserver3"`
	NameServer4 string `json:"nameserver4"`
	NameServer5 string `json:"nameserver5"`
	NameServer6 string `json:"nameserver6"`
}

type easynameNameserveChange struct {
	easynameResponseData
	Nameservers map[string]string `json:"data"`
}

func hashEncodeString(s string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(s)))
	return base64.StdEncoding.EncodeToString([]byte(hash))
}

func (c *easynameProvider) request(method, uri string, body *bytes.Buffer, result easynameResponse) error {
	httpClient := http.Client{}
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-ApiKey", c.apikey)
	req.Header.Set("X-User-Authentication", c.apiauth)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	bodyString, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyString, &result)

	status := result.GetStatus()
	if status.Type != "success" && status.Type != "pending" {
		return fmt.Errorf("easyname error (%d): %s", status.Code, status.Message)
	}
	return nil
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

	er := easynameResponseData{}
	if err = c.request("POST", fmt.Sprintf("https://api.easyname.com/domain/%d/nameserverchange", domain), bytes.NewBuffer(body), &er); err != nil {
		return err
	}

	return nil
}

func (c *easynameProvider) getDomain(domain string) (easynameDomain, error) {
	if c.domains == nil {
		c.fetchDomainList()
	}

	d, ok := c.domains[domain]
	if !ok {
		return easynameDomain{}, fmt.Errorf("the domain %s was not found in the easyname account", domain)
	}
	return d, nil
}

func (c *easynameProvider) fetchDomainList() error {
	c.domains = map[string]easynameDomain{}
	domains := easynameDomainList{}
	if err := c.request("GET", "https://api.easyname.com/domain", &bytes.Buffer{}, &domains); err != nil {
		return err
	}

	for _, domain := range domains.Domains {
		c.domains[domain.Domain] = domain
	}

	return nil
}
