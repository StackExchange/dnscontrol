//Package namedotcom implements a registrar that uses the name.com api to set name servers. It will self register it's providers when imported.
package namedotcom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/StackExchange/dnscontrol/providers"
)

const defaultApiBase = "https://api.name.com/api"

type nameDotCom struct {
	APIUrl  string `json:"apiurl"`
	APIUser string `json:"apiuser"`
	APIKey  string `json:"apikey"`
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf)
}

func newDsp(conf map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf)
}

func newProvider(conf map[string]string) (*nameDotCom, error) {
	api := &nameDotCom{}
	api.APIUser, api.APIKey, api.APIUrl = conf["apiuser"], conf["apikey"], conf["apiurl"]
	if api.APIKey == "" || api.APIUser == "" {
		return nil, fmt.Errorf("Name.com apikey and apiuser must be provided.")
	}
	if api.APIUrl == "" {
		api.APIUrl = defaultApiBase
	}
	return api, nil
}

func init() {
	providers.RegisterRegistrarType("NAMEDOTCOM", newReg)
	providers.RegisterDomainServiceProviderType("NAMEDOTCOM", newDsp, providers.CanUseAlias, providers.CanUseSRV)
	// PTR records are not supported https://www.name.com/support/articles/205188508-Reverse-DNS-records (2017-05-08)
}

///
//various http helpers for interacting with api
///

func (n *nameDotCom) addAuth(r *http.Request) {
	r.Header.Add("Api-Username", n.APIUser)
	r.Header.Add("Api-Token", n.APIKey)
}

type apiResult struct {
	Result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"result"`
}

func (r *apiResult) getErr() error {
	if r == nil {
		return nil
	}
	if r.Result.Code != 100 {
		if r.Result.Message == "" {
			return fmt.Errorf("Unknown error from name.com")
		}
		return fmt.Errorf(r.Result.Message)
	}
	return nil
}

//perform http GET and unmarshal response json into target struct
func (n *nameDotCom) get(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	n.addAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// perform http POST, json marshalling the given data into the body
func (n *nameDotCom) post(url string, data interface{}) (*apiResult, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}
	n.addAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &apiResult{}
	if err = json.Unmarshal(text, result); err != nil {
		return nil, err
	}
	return result, nil
}
