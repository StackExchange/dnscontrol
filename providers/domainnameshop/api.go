package domainnameshop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

var rootAPIURI = "https://api.domeneshop.no/v0"

func (api *domainNameShopProvider) getDomains(domainName string) ([]domainResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, rootAPIURI+"/domains?domain="+domainName, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(api.Token, api.Secret)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	domainResp := make([]domainResponse, 1)
	// SUGGESTION: Is "make" needed here or will "var" work?  If var works, use it. It is easier to read.  If it breaks the .Decode(), ignore this suggestion and use "make".
	// SUGGESTION: Don't use a variable with the same name as a type. It's confusing to the reader.
	//var domainResp []domainResponse
	err = json.NewDecoder(resp.Body).Decode(&domainResp)
	if err != nil {
		return nil, err
	}

	if domainName != "" && domainName != domainResp[0].Domain {
		return nil, fmt.Errorf("invalid domain name: %q != %q", domainName, domainResp[0].Domain)
	}

	return domainResp, nil
}

func (api *domainNameShopProvider) getDomainID(domainName string) (string, error) {
	domainResp, err := api.getDomains(domainName)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(domainResp[0].ID), nil
}

func (api *domainNameShopProvider) getNS(domainName string) ([]string, error) {
	domainResp, err := api.getDomains(domainName)
	if err != nil {
		return nil, err
	}
	return domainResp[0].Nameservers, nil
}

func (api *domainNameShopProvider) getDNS(domainName string) ([]domainNameShopRecord, error) {
	domainID, err := api.getDomainID(domainName)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, rootAPIURI+"/domains/"+domainID+"/dns", nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(api.Token, api.Secret)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	domainResponse := make([]domainNameShopRecord, 1)
	// SUGGESTION: Is "make" needed here or will "var" work?  If var works, use it. It is easier to read.  If it breaks the .Decode(), ignore this suggestion and use "make".
	//var domainResponse []domainNameShopRecord
	err = json.NewDecoder(resp.Body).Decode(&domainResponse)
	if err != nil {
		return nil, err
	}

	// SUGGESTION: Add a comment explaining what this loop does.
	// What I think it does is: Post-process the response. Mostly this involves
	// converting various strings to int (or uint16, etc) for future use.
	for i := range domainResponse {
		// Fix priority
		record := &domainResponse[i]
		priority, err := strconv.ParseUint(record.Priority, 10, 16)
		if err != nil {
			record.ActualPriority = 0
		}
		record.ActualPriority = uint16(priority)

		port, err := strconv.ParseUint(record.Port, 10, 16)
		if err != nil {
			record.ActualPort = 0
		}
		record.ActualPort = uint16(port)

		weight, err := strconv.ParseUint(record.Weight, 10, 16)
		if err != nil {
			record.ActualWeight = 0
		}
		record.ActualWeight = uint16(weight)

		// Fix CAA flags
		if record.Type == "CAA" {
			CaaFlag, err := strconv.ParseUint(record.ActualCAAFlag, 10, 8)
			if err != nil {
				record.CAAFlag = 0
			}
			record.CAAFlag = CaaFlag
		}

		// Transform data field to punycode if CNAME
		if record.Type == "CNAME" {
			punycodeData, err := idna.ToASCII(record.Data)
			if err != nil {
				return nil, err
			}
			record.Data = punycodeData
			if !strings.HasSuffix(record.Data, ".") {
				record.Data += "."
			}
		}

		record.TTL = uint16(fixTTL(uint32(record.TTL)))

		// Add domain id
		(&domainResponse[i]).DomainID = domainID
	}

	ns, err := api.getNS(domainName)
	if err != nil {
		return nil, err
	}

	// Adds NS as records
	for _, nameserver := range ns {
		domainResponse = append(domainResponse, domainNameShopRecord{
			ID:       0,
			Host:     "@",
			TTL:      300,
			Type:     "NS",
			Data:     nameserver + ".",
			DomainID: domainID,
		})
	}

	return domainResponse, nil
}

func (api *domainNameShopProvider) deleteRecord(domainID string, recordID string) error {
	return api.sendChangeRequest(http.MethodDelete, rootAPIURI+"/domains/"+domainID+"/dns/"+recordID, nil)
}

func (api *domainNameShopProvider) CreateRecord(domainName string, dnsR *domainNameShopRecord) error {
	domainID, err := api.getDomainID(domainName)
	if err != nil {
		return err
	}

	payloadBuf := new(bytes.Buffer)
	err = json.NewEncoder(payloadBuf).Encode(&dnsR)
	if err != nil {
		return err
	}

	return api.sendChangeRequest(http.MethodPost, rootAPIURI+"/domains/"+domainID+"/dns", payloadBuf)
}

func (api *domainNameShopProvider) UpdateRecord(dnsR *domainNameShopRecord) error {
	domainID := dnsR.DomainID
	recordID := strconv.Itoa(dnsR.ID)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(&dnsR)

	return api.sendChangeRequest(http.MethodPut, rootAPIURI+"/domains/"+domainID+"/dns/"+recordID, payloadBuf)
}

func (api *domainNameShopProvider) sendChangeRequest(method string, uri string, payload *bytes.Buffer) error {
	client := &http.Client{}

	var req *http.Request
	var err error
	if payload != nil {
		req, err = http.NewRequest(method, uri, payload)
	} else {
		req, err = http.NewRequest(method, uri, nil)
	}

	if err != nil {
		return err
	}

	req.SetBasicAuth(api.Token, api.Secret)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		// Record is deleted
		return nil
	case 204:
		//Update successful
		return nil
	case 400:
		return fmt.Errorf("DNS record failed validation")
	case 403:
		return fmt.Errorf("not authorized")
	case 404:
		return fmt.Errorf("does not exist")
	case 409:
		return fmt.Errorf("collision")
	default:
		return fmt.Errorf("unknown statuscode: %v", resp.StatusCode)
	}
}
