package linode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	mediaType      = "application/json"
	defaultBaseURL = "https://api.linode.com/v4/"
	domainsPath    = "domains"
)

func (c *LinodeAPI) fetchDomainList() error {
	c.domainIndex = map[string]int{}
	page := 1
	for {
		dr := &domainResponse{}
		endpoint := fmt.Sprintf("%s?page=%d", domainsPath, page)
		if err := c.get(endpoint, dr); err != nil {
			return fmt.Errorf("Error fetching domain list from Linode: %s", err)
		}
		for _, domain := range dr.Data {
			c.domainIndex[domain.Domain] = domain.ID
		}
		if len(dr.Data) == 0 || dr.Page >= dr.Pages {
			break
		}
		page++
	}
	return nil
}

func (c *LinodeAPI) getRecords(id int) ([]domainRecord, error) {
	records := []domainRecord{}
	page := 1
	for {
		dr := &recordResponse{}
		endpoint := fmt.Sprintf("%s/%d/records?page=%d", domainsPath, id, page)
		if err := c.get(endpoint, dr); err != nil {
			return nil, fmt.Errorf("Error fetching record list from Linode: %s", err)
		}

		records = append(records, dr.Data...)

		if len(dr.Data) == 0 || dr.Page >= dr.Pages {
			break
		}
		page++
	}

	return records, nil
}

func (c *LinodeAPI) createRecord(domainID int, rec *recordEditRequest) (*domainRecord, error) {
	endpoint := fmt.Sprintf("%s/%d/records", domainsPath, domainID)

	req, err := c.newRequest(http.MethodPost, endpoint, rec)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrors(resp)
	}

	record := &domainRecord{}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(record); err != nil {
		return nil, err
	}

	return record, nil
}

func (c *LinodeAPI) modifyRecord(domainID, recordID int, rec *recordEditRequest) error {
	endpoint := fmt.Sprintf("%s/%d/records/%d", domainsPath, domainID, recordID)

	req, err := c.newRequest(http.MethodPut, endpoint, rec)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrors(resp)
	}

	return nil
}

func (c *LinodeAPI) deleteRecord(domainID, recordID int) error {
	endpoint := fmt.Sprintf("%s/%d/records/%d", domainsPath, domainID, recordID)
	req, err := c.newRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrors(resp)
	}

	return nil
}

func (c *LinodeAPI) newRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	return req, nil
}

func (c *LinodeAPI) get(endpoint string, target interface{}) error {
	req, err := c.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return c.handleErrors(resp)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(target)
}

func (c *LinodeAPI) handleErrors(resp *http.Response) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	errs := &errorResponse{}

	if err := decoder.Decode(errs); err != nil {
		return fmt.Errorf("bad status code from Linode: %d not 200. Failed to decode response", resp.StatusCode)
	}

	buf := bytes.NewBufferString(fmt.Sprintf("bad status code from Linode: %d not 200", resp.StatusCode))

	for _, err := range errs.Errors {
		buf.WriteString("\n- ")

		if err.Field != "" {
			buf.WriteString(err.Field)
			buf.WriteString(": ")
		}

		buf.WriteString(err.Reason)
	}

	return errors.New(buf.String())
}

type basicResponse struct {
	Results int `json:"results"`
	Pages   int `json:"pages"`
	Page    int `json:"page"`
}

type domainResponse struct {
	basicResponse
	Data []struct {
		ID     int    `json:"id"`
		Domain string `json:"domain"`
	} `json:"data"`
}

type recordResponse struct {
	basicResponse
	Data []domainRecord `json:"data"`
}

type domainRecord struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Target   string `json:"target"`
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Service  string `json:"service"`
	Protocol string `json:"protocol"`
	TTLSec   uint32 `json:"ttl_sec"`
}

type recordEditRequest struct {
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Target   string `json:"target,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Port     int    `json:"port,omitempty"`
	Service  string `json:"service,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	// Documented as field `ttl` in the documentation, but in reality `ttl_sec` should be used
	TTL int `json:"ttl_sec,omitempty"`
}

type errorResponse struct {
	Errors []struct {
		Field  string `json:"field"`
		Reason string `json:"reason"`
	} `json:"errors"`
}
