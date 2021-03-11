package dnsme

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

const (
	baseUrlV2_0             = "https://api.dnsmadeeasy.com/V2.0/"
	sandboxBaseUrlV2_0      = "https://api.sandbox.dnsmadeeasy.com/V2.0/"
	requestDateHeaderLayout = "Mon, 2 Jan 2006 15:04:05 MST"
	initialBackoff          = time.Second * 10 // First backoff delay duration
	maxBackoff              = time.Minute * 3  // Maximum backoff delay
)

type dnsmeRestApi struct {
	baseUrl    string
	httpClient *http.Client

	apiKey    string
	secretKey string

	dumpHttpRequest  bool
	dumpHttpResponse bool
}

type apiErrorResponse struct {
	Error []string `json:"error"`
}

type apiEmptyResponse struct {
}

type apiRequest struct {
	method   string
	endpoint string
	data     []byte
}

func (restApi *dnsmeRestApi) singleDomainGet(domainId int) (*singleDomainResponse, error) {
	req := &apiRequest{
		method:   "GET",
		endpoint: fmt.Sprintf("dns/managed/%d", domainId),
	}

	res := &singleDomainResponse{}
	_, err := restApi.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (restApi *dnsmeRestApi) multiDomainGet() (*multiDomainResponse, error) {
	req := &apiRequest{
		method:   "GET",
		endpoint: "dns/managed/",
	}

	res := &multiDomainResponse{}
	_, err := restApi.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (restApi *dnsmeRestApi) recordGet(domainId int) (*recordResponse, error) {
	req := &apiRequest{
		method:   "GET",
		endpoint: fmt.Sprintf("dns/managed/%d/records", domainId),
	}

	res := &recordResponse{}
	_, err := restApi.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (restApi *dnsmeRestApi) singleDomainCreate(data singleDomainRequestData) (*singleDomainResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req := &apiRequest{
		method:   "POST",
		endpoint: fmt.Sprintf("dns/managed/"),
		data:     jsonData,
	}

	res := &singleDomainResponse{}
	_, err = restApi.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (restApi *dnsmeRestApi) recordDelete(domainId int, recordId int) error {
	req := &apiRequest{
		method:   "DELETE",
		endpoint: fmt.Sprintf("dns/managed/%d/records/%d", domainId, recordId),
	}

	_, err := restApi.sendRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (restApi *dnsmeRestApi) multiRecordCreate(domainId int, data []recordRequestData) (*[]recordResponseDataEntry, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req := &apiRequest{
		method:   "POST",
		endpoint: fmt.Sprintf("dns/managed/%d/records/createMulti", domainId),
		data:     jsonData,
	}

	res := &[]recordResponseDataEntry{}
	_, err = restApi.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (restApi *dnsmeRestApi) multiRecordUpdate(domainId int, data []recordRequestData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req := &apiRequest{
		method:   "PUT",
		endpoint: fmt.Sprintf("dns/managed/%d/records/updateMulti", domainId),
		data:     jsonData,
	}

	_, err = restApi.sendRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (restApi *dnsmeRestApi) multiRecordDelete(domainId int, recordIds []int) error {
	params := []string{}
	for i := range recordIds {
		params = append(params, fmt.Sprintf("ids=%d", recordIds[i]))
	}

	req := &apiRequest{
		method:   "DELETE",
		endpoint: fmt.Sprintf("dns/managed/%d/records?%s", domainId, strings.Join(params, "&")),
	}

	_, err := restApi.sendRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (restApi *dnsmeRestApi) createRequestAuthHeaders() (string, string) {
	t := time.Now()
	requestDate := t.UTC().Format(requestDateHeaderLayout)

	mac := hmac.New(sha1.New, []byte(restApi.secretKey))
	mac.Write([]byte(requestDate))
	macStr := hex.EncodeToString(mac.Sum(nil))

	return requestDate, macStr
}

func (restApi *dnsmeRestApi) createRequest(request *apiRequest) (*http.Request, error) {
	url := restApi.baseUrl + request.endpoint
	var req *http.Request
	var err error

	if request.method == "PUT" || request.method == "POST" {
		req, err = http.NewRequest(request.method, url, bytes.NewBuffer([]byte(request.data)))
	} else if request.method == "GET" || request.method == "DELETE" {
		req, err = http.NewRequest(request.method, url, nil)
	} else {
		return nil, fmt.Errorf("unknown API request method in DNSME REST API: %s", request.method)
	}

	if err != nil {
		return nil, err
	}

	requestDate, hmac := restApi.createRequestAuthHeaders()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-dnsme-apiKey", restApi.apiKey)
	req.Header.Set("x-dnsme-hmac", hmac)
	req.Header.Set("x-dnsme-requestDate", requestDate)

	return req, nil
}

// DNS Made Simple only allows 150 request / 5 minutes
// backoff is the amount of time to sleep if a "Rate limit exceeded" error is received
// It is increased up to maxBackoff after each use
// It is reset after successful request
var backoff = initialBackoff

func (restApi *dnsmeRestApi) sendRequest(request *apiRequest, response interface{}) (int, error) {
retry:
	req, err := restApi.createRequest(request)
	if err != nil {
		return 0, err
	}

	if restApi.dumpHttpRequest {
		dump, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dump))
	}

	res, err := restApi.httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()

	if restApi.dumpHttpResponse {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(dump))
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var apiErr apiErrorResponse
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			return res.StatusCode, fmt.Errorf("unknown error from DNSME REST API, status code: %d", res.StatusCode)
		}

		if len(apiErr.Error) == 1 && apiErr.Error[0] == "Rate limit exceeded" {
			fmt.Printf("pausing DNSME due to ratelimit: %v seconds\n", backoff)

			time.Sleep(backoff)

			backoff = backoff + (backoff / 2)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			goto retry
		}

		return res.StatusCode, fmt.Errorf("DNSME API error: %s", strings.Join(apiErr.Error, " "))
	}

	backoff = initialBackoff

	if response != nil {
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			return res.StatusCode, err
		}
	}

	return res.StatusCode, nil
}
