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

// LIVEAPI is the URL to the DNS Made Easy live (production) API. To use this you will need an account with DNS Made Easy (https://cp.dnsmadeeasy.com/)
const LIVEAPI = "https://api.dnsmadeeasy.com/V2.0/"

// SANDBOXAPI is the URL to the DNS Made Easy sandbox (testing) API. To use this you will need an account on the Sandbox system (https://sandbox.dnsmadeeasy.com/)
const SANDBOXAPI = "https://api.sandbox.dnsmadeeasy.com/V2.0/"

const pendingDeleteError = "Cannot delete a domain that is pending a create or delete action."

// GoDMEConfig is our struct that contains our API settings, client, etc
type GoDMEConfig struct {
	// APIUrl is the full URL of the API to use when communicating to DNS Made Easy. If omitted, this defaults to https://api.dnsmadeeasy.com/V2.0/
	APIUrl string
	// APIKey is your DNS Made Easy API key that can be obtained from https://dnsmadeeasy.com/account/info
	APIKey string
	// SecretKey is your DNS Made Easy API secret key that can be obtained from https://dnsmadeeasy.com/account/info
	SecretKey string
	// DisableSSLValidation disables the validation of the SSL certificate when using HTTPS. This is useful for the DNS Made Easy sandbox, which does not contain a valid certificate
	DisableSSLValidation bool
	// TimeAdjust is used for changing how fast/slow the timestamps used when authenticating with DNS Made Easy are. Normally you would just leave this at 0
	// and send a real timestamp, but DNS Made Easy has very strict requirements around time synchronisation. So if you're unlucky and your system time is a
	// touch fast or slow, you can adjust the timestamp we send using TimeAdjust to make it more accurate to UTC.
	TimeAdjust time.Duration
	dmeClient  *http.Client
}

// NewGoDNSMadeEasy must be called to construct a GoDMEConfig struct, otherwise there are uninitialised fields that may stop the API from working as expected
func NewGoDNSMadeEasy(dme *GoDMEConfig) (*GoDMEConfig, error) {
	if dme.APIKey == "" {
		return nil, fmt.Errorf("DNS Made Easy API key is blank")
	}

	if dme.SecretKey == "" {
		return nil, fmt.Errorf("DNS Made Easy API secret key is blank")
	}

	//If no API URL is specified, then default to the production API
	if dme.APIUrl == "" {
		dme.APIUrl = LIVEAPI
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

func (dme *GoDMEConfig) newRequest(Method, APIEndpoint string, body io.Reader) (*http.Request, error) {
	//Double check we have an API endpoint, just in case someone decides to create this object manually instead of
	//using NewGoDNSMadeEasy, or they screw around with it after it's created
	if dme.APIUrl == "" {
		dme.APIUrl = LIVEAPI
	}

	//Generate our Hex encoded HMAC SHA1 signature of the current date/time in UTC for our requests
	timeNow := time.Now().UTC()
	timeNow = timeNow.Add(dme.TimeAdjust)
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

func (dme *GoDMEConfig) doDMERequest(req *http.Request, dst interface{}) error {
	resp, err := dme.dmeClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if body != nil {
		//This is a stupid fix, because DNS Made Easy does not produce valid JSON for some of its error messages.
		body = []byte(strings.Replace(string(body), "{error:", "{\"error\":", 1))
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Access forbidden (%s)", req.URL.String())
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("404 Not Found (%s)", req.URL.String())
	}

	//fmt.Println(string(body))
	genericError := &GenericError{}

	//Try to unmarshal into an error to see if we get any data. A successful delete or update sends no body, so it might throw an error for DELETE or PUT, but that's OK
	json.Unmarshal(body, genericError)
	if len(genericError.Error) > 0 {
		return fmt.Errorf(strings.Join(genericError.Error, "\n"))
	}

	//If we are deleting a record and got this far, then it's been successful
	if req.Method == "PUT" || req.Method == "DELETE" {
		return nil
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
	Name                string        `json:"name"`
	NameServer          []string      `json:"nameServer,omitempty"`
	GtdEnabled          bool          `json:"gtdEnabled,omitempty"`
	ID                  int           `json:"id,omitempty"`
	FolderID            int           `json:"folderId,omitempty"`
	NameServers         []NameServer  `json:"nameServers"`
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

// NameServer is a DNS Made Easy Nameserver record
type NameServer struct {
	Fqdn string `json:"fqdn"`
	Ipv6 string `json:"ipv6"`
	Ipv4 string `json:"ipv4"`
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

// IPSet is a DNS Made Easy IP Set, used for secondary DNS
type IPSet struct {
	Name string   `json:"name"`
	ID   int      `json:"id"`
	Ips  []string `json:"ips"`
}

// SecondaryDomain is the configuration for a secondary DNS domain
type SecondaryDomain struct {
	Name              string       `json:"name"`
	ID                int          `json:"id"`
	FolderID          int          `json:"folderId"`
	NameServers       []NameServer `json:"nameServers,omitempty"`
	NameServerGroupID int          `json:"nameServerGroupId"`
	PendingActionID   int          `json:"pendingActionId,omitempty"`
	GtdEnabled        bool         `json:"gtdEnabled"`
	Updated           int64        `json:"updated,omitempty"`
	IPSet             IPSet        `json:"ipSet,omitempty"`
	IPSetID           int          `json:"ipSetId"`
	Created           int64        `json:"created,omitempty"`
}

// DomainExport is all of the data about a given domain that we can get from DNS Made Easy
type DomainExport struct {
	SOA       *SOA
	Info      *Domain
	DefaultNS *Vanity
	Records   *[]Record
}

// Folder is a DNS Made Easy folder, used in the list of folders.
type Folder struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

// FolderDetail is the detailed information regarding a folder
type FolderDetail struct {
	Name              string             `json:"name"`
	ID                int                `json:"id"`
	Domains           []int              `json:"domains,omitempty"`
	Secondaries       []int              `json:"secondaries,omitempty"`
	FolderPermissions []FolderPermission `json:"folderPermissions,omitempty"`
	DefaultFolder     bool               `json:"defaultFolder"`
}

// FolderPermission is used in the Folder struct
type FolderPermission struct {
	Permission int    `json:"permission"`
	FolderID   int    `json:"folderId"`
	GroupID    int    `json:"groupId"`
	FolderName string `json:"folderName"`
	GroupName  string `json:"groupName"`
}

// AllDomainExport is populated with all of the domains for a given account, and its records
type AllDomainExport map[string]DomainExport
