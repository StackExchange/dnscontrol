package GoDNSMadeEasy

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// GoDNSMadeEasy is our struct that contains our API settings, client, etc
type GoDNSMadeEasy struct {
	// APIUrl is the full URL of the API to use when communicating to DNS Made Easy. For example, https://api.dnsmadeeasy.com/V2.0/
	APIUrl string
	// APIKey is your DNS Made Easy API key that can be obtained from https://dnsmadeeasy.com/account/info
	APIKey string
	// SecretKey is your DNS Made Easy API secret key that can be obtained from https://dnsmadeeasy.com/account/info
	SecretKey string
	// DisableSSLValidation disables the validation of the SSL certificate when using HTTPS. This is useful for the DNS Made Easy sandbox, which does not contain a valid certificate
	DisableSSLValidation bool
	dmeClient            *http.Client
}

// NewGoDNSMadeEasy must be called to construct a GoDNSMadeEasy struct, otherwise there are uninitialised fields that may stop the API from working as expected
func NewGoDNSMadeEasy(dme *GoDNSMadeEasy) (*GoDNSMadeEasy, error) {
	if dme.APIKey == "" {
		return nil, fmt.Errorf("DNS Made Easy API key is blank")
	}

	if dme.SecretKey == "" {
		return nil, fmt.Errorf("DNS Made Easy API secret key is blank")
	}

	if dme.APIUrl == "" {
		return nil, fmt.Errorf("DNS Made Easy API URL is blank")
	}

	if string(dme.APIUrl[len(dme.APIUrl)-1]) != "/" {
		dme.APIUrl += "/"
	}

	//Create a HTTP transport that verifies SSL based on the user supplied parameter
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: dme.DisableSSLValidation,
		},
	}

	//Assign that transport to our new HTTP client (which we will reuse for all of the API requests)
	dme.dmeClient = &http.Client{Transport: tr}

	return dme, nil
}

func (dme *GoDNSMadeEasy) newRequest(Method, APIEndpoint string, body io.Reader) (*http.Request, error) {
	//Generate our Hex encoded HMAC SHA1 signature of the current date/time in UTC for our requests
	timeNow := time.Now().UTC()
	timeNow = timeNow.Add(15 * time.Second)
	timeNowString := timeNow.Format(time.RFC1123)
	key := []byte(dme.SecretKey)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(timeNowString))
	hmacSha := hex.EncodeToString(h.Sum(nil))

	thisRequestURI := dme.APIUrl + APIEndpoint
	thisReq, err := http.NewRequest(Method, thisRequestURI, body)
	if err != nil {
		return nil, err
	}
	thisReq.Header.Set("x-dnsme-apiKey", dme.APIKey)
	thisReq.Header.Set("x-dnsme-requestDate", timeNowString)
	thisReq.Header.Set("x-dnsme-hmac", hmacSha)
	thisReq.Header.Set("accept", "application/json")

	return thisReq, nil
}

func (dme *GoDNSMadeEasy) doDMERequest(req *http.Request, dst interface{}) error {
	resp, err := dme.dmeClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Access forbidden (%s)", req.URL.String())
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("404 Not Found (%s)", req.URL.String())
	}

	//fmt.Println(string(body))
	genericError := &GenericError{}

	//If we are deleting a record and got this far, then it's been successful
	if req.Method == "DELETE" {
		return nil
	}
	//Try to unmarshal into an error to see if we get any data
	err = json.Unmarshal(body, genericError)
	if err != nil {
		return fmt.Errorf("Could not parse response: %v\nData: %v", err, string(body))
	}
	if len(genericError.Error) > 0 {
		return fmt.Errorf(strings.Join(genericError.Error, "\n"))
	}

	err = json.Unmarshal(body, dst)
	return err //Will be null if unmarshals OK
}

// GenericResponse is a wrapper for the DNS Made Easy responses. All the useful information is in the Data field.
type GenericResponse struct {
	Page         int `json:"page"`
	TotalPages   int `json:"totalPages"`
	TotalRecords int `json:"totalRecords"`
	Data         json.RawMessage
}

// GenericError contains a generic array of strings that represent errors when interacting with the API
type GenericError struct {
	Error []string `json:"error"`
}

// Domain is our basic information regarding a domain. This does not contain any records.
type Domain struct {
	Name        string   `json:"name"`
	NameServer  []string `json:"nameServer,omitempty"`
	GtdEnabled  bool     `json:"gtdEnabled,omitempty"`
	ID          int      `json:"id,omitempty"`
	FolderID    int      `json:"folderId,omitempty"`
	NameServers []struct {
		Fqdn string `json:"fqdn"`
		Ipv6 string `json:"ipv6"`
		Ipv4 string `json:"ipv4"`
	} `json:"nameServers"`
	Updated             int64         `json:"updated,omitempty"`
	TemplateID          int           `json:"templateId,omitempty"`
	DelegateNameServers []string      `json:"delegateNameServers,omitempty"`
	Created             int64         `json:"created,omitempty"`
	TransferAclID       int           `json:"transferAclId,omitempty"`
	ActiveThirdParties  []interface{} `json:"activeThirdParties,omitempty"`
	VanityID            int           `json:"vanityId,omitempty"`
	PendingActionID     int           `json:"pendingActionId,omitempty"`
	SoaID               int           `json:"soaId,omitempty"`
	ProcessMulti        bool          `json:"processMulti,omitempty"`
}

// Record represents a DNS record from DNS Made Easy (e.g. A, AAAA, PTR, NS, etc)
type Record struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	ID           int    `json:"id"`
	Type         string `json:"type"`
	DynamicDNS   bool   `json:"dynamicDns"`
	Failed       bool   `json:"failed"`
	GtdLocation  string `json:"gtdLocation"`
	HardLink     bool   `json:"hardLink"`
	TTL          int    `json:"ttl"`
	Failover     bool   `json:"failover"`
	Monitor      bool   `json:"monitor"`
	SourceID     int    `json:"sourceId"`
	Source       int    `json:"source"`
	MxLevel      int    `json:"mxLevel,omitempty"`
	Priority     int    `json:"priority,omitempty"`
	Port         int    `json:"port,omitempty"`
	Weight       int    `json:"weight,omitempty"`
	Keywords     string `json:"keywords,omitempty"`
	RedirectType string `json:"redirectType,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
}

// SOA represents a Start of Authority configuration from DNS Made Easy
type SOA struct {
	Name          string `json:"name"`
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Comp          string `json:"comp"`
	Refresh       int    `json:"refresh"`
	Serial        int    `json:"serial"`
	Retry         int    `json:"retry"`
	Expire        int    `json:"expire"`
	NegativeCache int    `json:"negativeCache"`
	TTL           int    `json:"ttl"`
}

// Vanity represents a vanity nameserver configuration from DNS Made Easy
type Vanity struct {
	Name              string   `json:"name"`
	ID                int      `json:"id"`
	NameServerGroupID int      `json:"nameServerGroupId"`
	NameServerGroup   string   `json:"nameServerGroup"`
	Servers           []string `json:"servers"`
	Public            bool     `json:"public"`
	Default           bool     `json:"default"`
}

// DomainExport is all of the data about a given domain that we can get from DNS Made Easy
type DomainExport struct {
	SOA       *SOA
	Info      *Domain
	DefaultNS *Vanity
	Records   *[]Record
}

// AllDomainExport is populated with all of the domains for a given account, and its records
type AllDomainExport map[string]DomainExport
