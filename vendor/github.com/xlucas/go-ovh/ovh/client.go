package ovh

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// OVH endpoints list
const (
	ENDPOINT_CA_OVHCOM     = "https://ca.api.ovh.com/1.0"
	ENDPOINT_CA_KIMSUFI    = "https://ca.api.kimsufi.com/1.0"
	ENDPOINT_CA_RUNABOVE   = "https://api.runabove.com/1.0"
	ENDPOINT_CA_SOYOUSTART = "https://ca.api.soyoustart.com/1.0"
	ENDPOINT_EU_OVHCOM     = "https://eu.api.ovh.com/1.0"
	ENDPOINT_EU_KIMSUFI    = "https://eu.api.kimsufi.com/1.0"
	ENDPOINT_EU_RUNABOVE   = "https://api.runabove.com/1.0"
	ENDPOINT_EU_SOYOUSTART = "https://eu.api.soyoustart.com/1.0"
)

// Client helps interacting with OVH API endpoints.
type Client struct {
	AppKey      string
	AppSecret   string
	ConsumerKey string
	Endpoint    string
	TimeShift   time.Duration
	Debug       bool
}

// NewClient builds up a new client link to the specified endpoint
// with given authentication information and no timeshift.
func NewClient(endpoint, ak, as, ck string, debug bool) *Client {
	return &Client{
		AppKey:      ak,
		AppSecret:   as,
		ConsumerKey: ck,
		Endpoint:    endpoint,
		TimeShift:   0,
		Debug:       debug,
	}
}

func computeSignature(appSecret, consumerKey, method, url string, body []byte, timestamp int64) string {
	hasher := sha1.New()
	pattern := fmt.Sprintf("%s+%s+%s+%s+%s+%d",
		appSecret,
		consumerKey,
		method,
		url,
		body,
		timestamp)
	hasher.Write([]byte(pattern))
	return fmt.Sprintf("$1$%x", hasher.Sum(nil))
}

func sendRequest(appKey, consumerKey, signature string, timestamp int64, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Ovh-Application", appKey)
	req.Header.Add("X-Ovh-Consumer", consumerKey)
	req.Header.Add("X-Ovh-Signature", signature)
	req.Header.Add("X-Ovh-Timestamp", fmt.Sprintf("%d", timestamp))

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	outBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Unexpected HTTP return code (%s : %s).", resp.Status, outBytes)
	}

	return outBytes, err
}

// PollTimeshift calculates the difference between
// local and remote system time through a call to
// the API. It may be useful to call this function
// to avoid signatures to be rejected due to
// timeshift or network delay.
func (c *Client) PollTimeshift() error {
	sysTime := time.Now()
	resp, err := http.Get(c.Endpoint + "/auth/time")
	if err != nil {
		return err
	}
	outPayload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	apiTime, err := strconv.ParseInt(string(outPayload), 10, 64)
	if err != nil {
		return err
	}
	c.TimeShift = time.Unix(apiTime, 0).Sub(sysTime)
	return err
}

// Call sends a request to the OVH API and returns response content.
// Input and output json processing will leverage json
// marshalling/unmarshalling of the specified interfaces.
func (c *Client) Call(method, path string, in interface{}, out interface{}) error {
	var (
		inBytes, outBytes []byte
		err               error
	)
	if in != nil {
		inBytes, err = json.Marshal(in)
	}
	if err != nil {
		return err
	}

	url := c.Endpoint + path
	timestamp := time.Now().Add(c.TimeShift).Unix()
	signature := computeSignature(c.AppSecret, c.ConsumerKey, method, url, inBytes, timestamp)

	if c.Debug {
		log.Printf("Method = %s", method)
		log.Printf("URL = %s", url)
		log.Printf("Timestamp = %d", timestamp)
		log.Printf("Signature = %s", signature)
		log.Printf("Body = %s", inBytes)
	}

	outBytes, err = sendRequest(c.AppKey, c.ConsumerKey, signature, timestamp, method, url, inBytes)
	if err != nil {
		return err
	}

	err = json.Unmarshal(outBytes, &out)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SetDebug(debug bool) {
	c.Debug = debug
}
