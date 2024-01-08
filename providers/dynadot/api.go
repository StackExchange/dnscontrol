package dynadot

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// API layer for Dynadot

type dynadotProvider struct {
	key string
}

type requestParams map[string]string

type header struct {
	SuccessCode int    `xml:"SuccessCode"`
	Status      string `xml:"Status"`
	Error       string `xml:"Error,omitempty"`
}

type addNsResponse struct {
	XMLName     xml.Name `xml:"AddNsResponse"`
	AddNsHeader header   `xml:"AddNsHeader"`
}

type setNsResponse struct {
	XMLName     xml.Name `xml:"SetNsResponse"`
	SetNsHeader header   `xml:"SetNsHeader"`
}

type getNsResponse struct {
	XMLName     xml.Name  `xml:"GetNsResponse"`
	NsContent   nsContent `xml:"NsContent"`
	GetNsHeader header    `xml:"GetNsHeader"`
}

type nsContent struct {
	Host   []string `xml:"Host"`
	NsName string   `xml:"NsName"`
}

func (c *dynadotProvider) getNameservers(domain string) ([]string, error) {
	var bodyString, err = c.get("get_ns", requestParams{"domain": domain})
	if err != nil {
		return []string{}, fmt.Errorf("failed NS list (Dynadot): %s", err)
	}
	var ns getNsResponse
	xml.Unmarshal(bodyString, &ns)

	if ns.GetNsHeader.SuccessCode != 0 {
		return []string{}, fmt.Errorf("failed NS list (Dynadot): %s", ns.GetNsHeader.Error)
	}

	hosts := []string{}
	hosts = append(hosts, ns.NsContent.Host...)
	return hosts, nil
}

func (c *dynadotProvider) updateNameservers(ns []string, domain string) error {
	if len(ns) > 13 {
		return fmt.Errorf("failed NS update (Dynadot): only up to 13 nameservers are supported")
	}

	// Nameservers must first be added to the Dynadot account
	for _, host := range ns {
		b, err := c.get("add_ns", requestParams{"host": host})
		if err != nil {
			return fmt.Errorf("failed NS add (Dynadot): %s", err)
		}
		var resp addNsResponse
		err = xml.Unmarshal(b, &resp)
		if err != nil {
			return fmt.Errorf("failed NS add (Dynadot): %s", err)
		}

		if resp.AddNsHeader.SuccessCode != 0 {
			// No apparent way to get all existing entries on an account, so filter
			if strings.Contains(resp.AddNsHeader.Error, "already exists") {
				continue
			}
			return fmt.Errorf("failed NS add (Dynadot): %s", resp.AddNsHeader.Error)

		}
	}

	rec := requestParams{}
	rec["domain"] = domain
	// supported prams: ns0 - ns12
	for i, h := range ns {
		rec[fmt.Sprintf("%s%d", "ns", i)] = h
	}

	b, err := c.get("set_ns", rec)
	if err != nil {
		return fmt.Errorf("failed NS set (Dynadot): %s", err)
	}

	var resp setNsResponse
	err = xml.Unmarshal(b, &resp)
	if err != nil {
		return fmt.Errorf("failed NS add (Dynadot): %s", err)
	}

	if resp.SetNsHeader.SuccessCode != 0 {
		return fmt.Errorf("failed NS add (Dynadot): %s", resp.SetNsHeader.Error)
	}

	return nil
}

func (c *dynadotProvider) get(command string, params requestParams) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.dynadot.com/api3.xml", nil)
	q := req.URL.Query()

	q.Add("key", c.key)
	q.Add("command", command)

	for pName, pValue := range params {
		q.Add(pName, pValue)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	return io.ReadAll(resp.Body)
}
