package internetbs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Api layer for Internet.bs

type internetbsProvider struct {
	key      string
	password string
}

type requestParams map[string]string

type errorResponse struct {
	TransactID string `json:"transactid"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Code       uint   `json:"code,omitempty"`
}

type domainRecord struct {
	Nameserver []string `json:"nameserver"`
}

func (c *internetbsProvider) getNameservers(domain string) ([]string, error) {
	var bodyString, err = c.get("/Domain/Info", requestParams{"Domain": domain})
	if err != nil {
		return []string{}, fmt.Errorf("failed fetching nameservers list (Internet.bs): %s", err)
	}
	var dr domainRecord
	json.Unmarshal(bodyString, &dr)
	ns := []string{}
	ns = append(ns, dr.Nameserver...)
	return ns, nil
}

func (c *internetbsProvider) updateNameservers(ns []string, domain string) error {
	rec := requestParams{}
	rec["Domain"] = domain
	rec["Ns_list"] = strings.Join(ns, ",")
	if _, err := c.get("/Domain/Update", rec); err != nil {
		return fmt.Errorf("failed NS update (Internet.bs): %s", err)
	}
	return nil
}

func (c *internetbsProvider) get(endpoint string, params requestParams) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.internet.bs/"+endpoint, nil)
	q := req.URL.Query()

	// Add auth params
	q.Add("ApiKey", c.key)
	q.Add("Password", c.password)
	q.Add("ResponseFormat", "JSON")

	for pName, pValue := range params {
		q.Add(pName, pValue)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)

	// Got error from API ?
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		return []byte{}, err
	}
	if errResp.Status == "FAILURE" {
		return bodyString, fmt.Errorf("failed API (Internet.bs): %s code: %d transactid: %s  URL:%s%s ",
			errResp.Message, errResp.Code, errResp.TransactID,
			req.Host, req.URL.RequestURI())
	}

	return bodyString, nil
}
