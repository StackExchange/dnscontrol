package desec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
)

const apiBase = "https://desec.io/api/v1"

// Api layer for desec
type desecProvider struct {
	domainIndex map[string]uint32 //stores the minimum ttl of each domain. (key = domain and value = ttl)
	creds       struct {
		tokenid  string
		token    string
		user     string
		password string
	}
	mutex sync.Mutex
}

type domainObject struct {
	Created    time.Time   `json:"created,omitempty"`
	Keys       []dnssecKey `json:"keys,omitempty"`
	MinimumTTL uint32      `json:"minimum_ttl,omitempty"`
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
type nonFieldError struct {
	Errors []string `json:"non_field_errors"`
}

func (c *desecProvider) authenticate() error {
	endpoint := "/domains/"
	var _, _, err = c.get(endpoint, "GET")
	if err != nil {
		return err
	}
	return nil
}
func (c *desecProvider) initializeDomainIndex() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.domainIndex != nil {
		return nil
	}
	endpoint := "/domains/"
	var bodyString, resp, err = c.get(endpoint, "GET")
	if resp.StatusCode == 400 && resp.Header.Get("Link") != "" {
		//pagination is required
		links := c.convertLinks(resp.Header.Get("Link"))
		endpoint = links["first"]
		printer.Debugf("initial endpoint %s\n", endpoint)
		for endpoint != "" {
			bodyString, resp, err = c.get(endpoint, "GET")
			if err != nil {
				if resp.StatusCode == 404 {
					return nil
				}
				return fmt.Errorf("failed fetching domains: %s", err)
			}
			err = c.buildIndexFromResponse(bodyString)
			if err != nil {
				return fmt.Errorf("failed fetching domains: %s", err)
			}
			links = c.convertLinks(resp.Header.Get("Link"))
			endpoint = links["next"]
			printer.Debugf("next endpoint %s\n", endpoint)
		}
		printer.Debugf("Domain Index initilized with pagination (%d domains)\n", len(c.domainIndex))
		return nil //domainIndex was build using pagination without errors
	}

	//no pagination required
	if err != nil && resp.StatusCode != 400 {
		if resp.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("failed fetching domains: %s", err)
	}
	err = c.buildIndexFromResponse(bodyString)
	if err == nil {
		printer.Debugf("Domain Index initilized without pagination (%d domains)\n", len(c.domainIndex))
	}
	return err
}

// buildIndexFromResponse takes the bodyString from initializeDomainIndex and builds the domainIndex
func (c *desecProvider) buildIndexFromResponse(bodyString []byte) error {
	if c.domainIndex == nil {
		c.domainIndex = map[string]uint32{}
	}
	var dr []domainObject
	err := json.Unmarshal(bodyString, &dr)
	if err != nil {
		return err
	}
	for _, domain := range dr {
		//deSEC allows different minimum ttls per domain
		//we store the actual minimum ttl to use it in desecProvider.go GetDomainCorrections() to enforce the minimum ttl and avoid api errors.
		c.domainIndex[domain.Name] = domain.MinimumTTL
	}
	return nil
}

// Parses the Link Header into a map (https://github.com/desec-io/desec-tools/blob/master/fetch_zone.py#L13)
func (c *desecProvider) convertLinks(links string) map[string]string {
	mapping := make(map[string]string)
	printer.Debugf("Header: %s\n", links)
	for _, link := range strings.Split(links, ", ") {
		tmpurl := strings.Split(link, "; ")
		if len(tmpurl) != 2 {
			printer.Printf("unexpected link header %s", link)
			continue
		}
		r := regexp.MustCompile(`rel="(.*)"`)
		matches := r.FindStringSubmatch(tmpurl[1])
		if len(matches) != 2 {
			printer.Printf("unexpected label %s", tmpurl[1])
			continue
		}
		// mapping["$label"] = "$URL"
		//URL = https://desec.io/api/v1/domains/{domain}/rrsets/?cursor=:next_cursor
		mapping[matches[1]] = strings.TrimSuffix(strings.TrimPrefix(tmpurl[0], "<"), ">")
	}
	return mapping
}

func (c *desecProvider) getRecords(domain string) ([]resourceRecord, error) {
	endpoint := "/domains/%s/rrsets/"
	var rrsNew []resourceRecord
	var bodyString, resp, err = c.get(fmt.Sprintf(endpoint, domain), "GET")
	if resp.StatusCode == 400 && resp.Header.Get("Link") != "" {
		//pagination required
		links := c.convertLinks(resp.Header.Get("Link"))
		endpoint = links["first"]
		printer.Debugf("getRecords: initial endpoint %s\n", fmt.Sprintf(endpoint, domain))
		for endpoint != "" {
			bodyString, resp, err = c.get(endpoint, "GET")
			if err != nil {
				if resp.StatusCode == 404 {
					return rrsNew, nil
				}
				return rrsNew, fmt.Errorf("getRecords: failed fetching rrsets: %s", err)
			}
			tmp, err := generateRRSETfromResponse(bodyString)
			if err != nil {
				return rrsNew, fmt.Errorf("failed fetching records for domain %s (deSEC): %s", domain, err)
			}
			rrsNew = append(rrsNew, tmp...)
			links = c.convertLinks(resp.Header.Get("Link"))
			endpoint = links["next"]
			printer.Debugf("getRecords: next endpoint %s\n", endpoint)
		}
		printer.Debugf("Build rrset using pagination (%d rrs)\n", len(rrsNew))
		return rrsNew, nil //domainIndex was build using pagination without errors
	}
	//no pagination
	if err != nil {
		return rrsNew, fmt.Errorf("failed fetching records for domain %s (deSEC): %s", domain, err)
	}
	tmp, err := generateRRSETfromResponse(bodyString)
	if err != nil {
		return rrsNew, err
	}
	rrsNew = append(rrsNew, tmp...)
	printer.Debugf("Build rrset without pagination (%d rrs)\n", len(rrsNew))
	return rrsNew, nil
}

// generateRRSETfromResponse takes the response rrset api calls and returns []resourceRecord
func generateRRSETfromResponse(bodyString []byte) ([]resourceRecord, error) {
	var rrs []rrResponse
	var rrsNew []resourceRecord
	err := json.Unmarshal(bodyString, &rrs)
	if err != nil {
		return rrsNew, err
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
		rrsNew = append(rrsNew, tmp)
	}
	return rrsNew, nil
}

func (c *desecProvider) createDomain(domain string) error {
	endpoint := "/domains/"
	pl := domainObject{Name: domain}
	byt, _ := json.Marshal(pl)
	var resp []byte
	var err error
	if resp, err = c.post(endpoint, "POST", byt); err != nil {
		return fmt.Errorf("failed domain create (deSEC): %v", err)
	}
	dm := domainObject{}
	err = json.Unmarshal(resp, &dm)
	if err != nil {
		return err
	}
	printer.Printf("To enable DNSSEC validation for your domain, make sure to convey the DS record(s) to your registrar:\n")
	printer.Printf("%+q", dm.Keys)
	return nil
}

// upsertRR will create or override the RRSet with the provided resource record.
func (c *desecProvider) upsertRR(rr []resourceRecord, domain string) error {
	endpoint := fmt.Sprintf("/domains/%s/rrsets/", domain)
	byt, _ := json.Marshal(rr)
	if _, err := c.post(endpoint, "PUT", byt); err != nil {
		return fmt.Errorf("failed create RRset (deSEC): %v", err)
	}
	return nil
}

func (c *desecProvider) deleteRR(domain, shortname, t string) error {
	endpoint := fmt.Sprintf("/domains/%s/rrsets/%s/%s/", domain, shortname, t)
	if _, _, err := c.get(endpoint, "DELETE"); err != nil {
		return fmt.Errorf("failed delete RRset (deSEC): %v", err)
	}
	return nil
}

func (c *desecProvider) get(target, method string) ([]byte, *http.Response, error) {
	retrycnt := 0
	var endpoint string
	if strings.Contains(target, "http") {
		endpoint = target
	} else {
		endpoint = apiBase + target
	}
retry:
	client := &http.Client{}
	req, _ := http.NewRequest(method, endpoint, nil)
	q := req.URL.Query()
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.creds.token))

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, resp, err
	}

	bodyString, _ := io.ReadAll(resp.Body)
	// Got error from API ?
	if resp.StatusCode > 299 {
		if resp.StatusCode == 429 && retrycnt < 5 {
			retrycnt++
			//we've got rate limiting and will try to get the Retry-After Header if this fails we fallback to sleep for 500ms max. 5 retries.
			waitfor := resp.Header.Get("Retry-After")
			if waitfor != "" {
				wait, err := strconv.ParseInt(waitfor, 10, 64)
				if err == nil {
					if wait > 180 {
						return []byte{}, resp, fmt.Errorf("rate limiting exceeded")
					}
					printer.Warnf("Rate limiting.. waiting for %s seconds", waitfor)
					time.Sleep(time.Duration(wait+1) * time.Second)
					goto retry
				}
			}
			printer.Warnf("Rate limiting.. waiting for 500 milliseconds")
			time.Sleep(500 * time.Millisecond)
			goto retry
		}
		var errResp errorResponse
		var nfieldErrors []nonFieldError
		err = json.Unmarshal(bodyString, &errResp)
		if err == nil {
			return bodyString, resp, fmt.Errorf("%s", errResp.Detail)
		}
		err = json.Unmarshal(bodyString, &nfieldErrors)
		if err == nil && len(nfieldErrors) > 0 {
			if len(nfieldErrors[0].Errors) > 0 {
				return bodyString, resp, fmt.Errorf("%s", nfieldErrors[0].Errors[0])
			}
		}
		return bodyString, resp, fmt.Errorf("HTTP status %s Body: %s, the API does not provide more information", resp.Status, bodyString)
	}
	return bodyString, resp, nil
}

func (c *desecProvider) post(target, method string, payload []byte) ([]byte, error) {
	retrycnt := 0
	var endpoint string
	if strings.Contains(target, "http") {
		endpoint = target
	} else {
		endpoint = apiBase + target
	}
retry:
	client := &http.Client{}
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(payload))
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

	bodyString, _ := io.ReadAll(resp.Body)

	// Got error from API ?
	if resp.StatusCode > 299 {
		if resp.StatusCode == 429 && retrycnt < 5 {
			retrycnt++
			//we've got rate limiting and will try to get the Retry-After Header if this fails we fallback to sleep for 500ms max. 5 retries.
			waitfor := resp.Header.Get("Retry-After")
			if waitfor != "" {
				wait, err := strconv.ParseInt(waitfor, 10, 64)
				if err == nil {
					if wait > 180 {
						return []byte{}, fmt.Errorf("rate limiting exceeded")
					}
					printer.Warnf("Rate limiting.. waiting for %s seconds", waitfor)
					time.Sleep(time.Duration(wait+1) * time.Second)
					goto retry
				}
			}
			printer.Warnf("Rate limiting.. waiting for 500 milliseconds")
			time.Sleep(500 * time.Millisecond)
			goto retry
		}
		var errResp errorResponse
		var nfieldErrors []nonFieldError
		err = json.Unmarshal(bodyString, &errResp)
		if err == nil {
			return bodyString, fmt.Errorf("HTTP status %d %s details: %s", resp.StatusCode, resp.Status, errResp.Detail)
		}
		err = json.Unmarshal(bodyString, &nfieldErrors)
		if err == nil && len(nfieldErrors) > 0 {
			if len(nfieldErrors[0].Errors) > 0 {
				return bodyString, fmt.Errorf("%s", nfieldErrors[0].Errors[0])
			}
		}
		return bodyString, fmt.Errorf("HTTP status %s Body: %s, the API does not provide more information", resp.Status, bodyString)
	}
	//time.Sleep(334 * time.Millisecond)
	return bodyString, nil
}
