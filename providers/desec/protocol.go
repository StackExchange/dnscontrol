package desec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Api layer for desec
type api struct {
	domainIndex      map[string]string
	nameserversNames []string
	creds            struct {
		tokenid  string
		token    string
		user     string
		password string
	}
}

type domainObject struct {
	Created    time.Time   `json:"created,omitempty"`
	Keys       []dnssecKey `json:"keys,omitempty"`
	MinimumTTL int         `json:"minimum_ttl,omitempty"`
	Name       string      `json:"name,omitempty"`
	Published  time.Time   `json:"published,omitempty"`
}

type resourceRecord struct {
	Subname string   `json:"subname"`
	Records []string `json:"records"`
	TTL     uint32   `json:"ttl,omitempty"`
	Type    string   `json:"type"`
	Target  string   `json:"-"`
}

type rrResponse struct {
	resourceRecord
	Created time.Time `json:"created"`
	Domain  string    `json:"domain"`
	Name    string    `json:"name"`
}

type dnssecKey struct {
	Dnskey  string   `json:"dnskey"`
	Ds      []string `json:"ds"`
	Flags   int      `json:"flags"`
	Keytype string   `json:"keytype"`
}

type errorResponse struct {
	Detail string `json:"detail"`
}

func (c *api) fetchDomainList() error {
	c.domainIndex = map[string]string{}
	var dr []domainObject
	endpoint := "/domains/"
	var bodyString, err = c.get(endpoint, "GET")
	if err != nil {
		return fmt.Errorf("Error fetching domain list from deSEC: %s", err)
	}
	err = json.Unmarshal(bodyString, &dr)
	if err != nil {
		return err
	}
	for _, domain := range dr {
		c.domainIndex[domain.Name] = domain.Name
	}
	return nil
}

func (c *api) getRecords(domain string) ([]resourceRecord, error) {
	endpoint := "/domains/%s/rrsets/"
	var rrs []rrResponse
	var rrs_new []resourceRecord
	var bodyString, err = c.get(fmt.Sprintf(endpoint, domain), "GET")
	if err != nil {
		return rrs_new, fmt.Errorf("Error fetching records from deSEC for domain %s: %s", domain, err)
	}
	err = json.Unmarshal(bodyString, &rrs)
	if err != nil {
		return rrs_new, err
	}
	// deSEC returns round robin records as array but dnsconfig expects single entries for each record
	// we will create one object per record except of TXT records which are handled as array of string by dnscontrol aswell.
	for i := range rrs {
		tmp := resourceRecord{
			TTL:     rrs[i].TTL,
			Type:    rrs[i].Type,
			Subname: rrs[i].Subname,
			Records: rrs[i].Records,
		}
		rrs_new = append(rrs_new, tmp)
	}
	return rrs_new, nil
}

func (c *api) createDomain(domain string) error {
	endpoint := "/domains/"
	pl := domainObject{Name: domain}
	byt, _ := json.Marshal(pl)
	if _, err := c.post(endpoint, "POST", byt); err != nil {
		return fmt.Errorf("Error create domain deSEC: %v", err)
	}
	return nil
}

//upsertRR will create or override the RRSet with the provided resource record.
func (c *api) upsertRR(rr resourceRecord, domain string) error {
	endpoint := fmt.Sprintf("/domains/%s/rrsets/", domain)
	var rrs []resourceRecord
	rrs = append(rrs, rr)
	byt, _ := json.Marshal(rrs)
	if _, err := c.post(endpoint, "PATCH", byt); err != nil {
		return fmt.Errorf("Error create rrset deSEC: %v", err)
	}
	return nil
}

func (c *api) deleteRR(domain, shortname, t string) error {
	endpoint := fmt.Sprintf("/domains/%s/rrsets/%s/%s/", domain, shortname, t)
	if _, err := c.get(endpoint, "DELETE"); err != nil {
		return fmt.Errorf("Error delete rrset deSEC: %v", err)
	}
	return nil
}

func (c *api) get(endpoint, method string) ([]byte, error) {
	retrycnt := 0
retry:
	client := &http.Client{}
	req, _ := http.NewRequest(method, "https://desec.io/api/v1"+endpoint, nil)
	q := req.URL.Query()
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.creds.token))

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	// Got error from API ?
	if resp.StatusCode > 299 {
		if resp.StatusCode == 429 && retrycnt < 5 {
			retrycnt++
			time.Sleep(500 * time.Millisecond)
			goto retry
		}
		var errResp errorResponse
		err = json.Unmarshal(bodyString, &errResp)
		if err == nil {
			return bodyString, fmt.Errorf("%s", errResp.Detail)
		}
		return bodyString, fmt.Errorf("http status %d %s, the api does not provide more information", resp.StatusCode, resp.Status)
	}
	return bodyString, nil
}

func (c *api) post(endpoint, method string, payload []byte) ([]byte, error) {
	retrycnt := 0
retry:
	client := &http.Client{}
	req, err := http.NewRequest(method, "https://desec.io/api/v1"+endpoint, bytes.NewReader(payload))
	if err != nil {
		return []byte{}, err
	}
	q := req.URL.Query()
	if endpoint != "/auth/login/" {
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.creds.token))
	}
	req.Header.Set("Content-Type", "application/json")

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)

	// Got error from API ?
	if resp.StatusCode > 299 {
		if resp.StatusCode == 429 && retrycnt < 5 {
			retrycnt++
			time.Sleep(500 * time.Millisecond)
			goto retry
		}
		var errResp errorResponse
		err = json.Unmarshal(bodyString, &errResp)
		if err == nil {
			return bodyString, fmt.Errorf("http status %d %s details: %s", resp.StatusCode, resp.Status, errResp.Detail)
		}
		return bodyString, fmt.Errorf("http status %d %s, the api does not provide more information", resp.StatusCode, resp.Status)
	}
	//time.Sleep(334 * time.Millisecond)
	return bodyString, nil
}
