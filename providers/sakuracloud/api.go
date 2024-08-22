// NOTE: As the API documentation of Sakura Cloud is written in Japanese
// and lacks further explanation, we have described the API data structures
// in English in the structure comments.
//
// - https://manual.sakura.ad.jp/cloud-api/1.1/appliance/index.html

package sakuracloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"time"
)

// requestCommonServiceItem is the body structure of the request to create a zone or update zone data.
//
// Zone creation:
//
//	POST /commonserviceitem
//
//	{
//	  "CommonServiceItem": {
//	    "Name": "example.com",
//	    "Status": {
//	      "Zone": "example.com"
//	    },
//	    "Settings": {
//	      "DNS": {
//	        "ResourceRecordSets": []
//	      }
//	    },
//	    "Provider": {
//	      "Class": "dns"
//	    },
//	  }
//	}
//
// Zone update:
//
//	PUT /commonserviceitem/:commonserviceitemid
//
//	{
//	  "CommonServiceItem": {
//	    "Settings": {
//	      "DNS": {
//	        "ResourceRecordSets": [
//	          {
//	            "Name": "a",
//	            "Type": "A",
//	            "RData": "192.0.2.1",
//	            "TTL": 600
//	          },
//	          ...
//	        ]
//	      }
//	    }
//	  }
//	}
//
// Reference:
//
//   - https://manual.sakura.ad.jp/cloud-api/1.1/appliance/#post_commonserviceitem
//   - https://manual.sakura.ad.jp/cloud-api/1.1/appliance/#put_commonserviceitem_commonserviceitemid
type requestCommonServiceItem struct {
	CommonServiceItem commonServiceItem `json:"CommonServiceItem"`
}

// responseCommonServiceItems is the body structure of the success response to get a list of zones.
//
// Request:
//
//	GET /commonserviceitem
//
// Response body structure:
//
//	{
//	  "From": 0,
//	  "Count": 1,
//	  "Total": 1,
//	  "CommonServiceItems": [
//	    {
//	      "Index": 0,
//	      "ID": "999999999999",
//	      "Name": "example.com",
//	      "Description": "",
//	      "Settings": {
//	        "DNS": {
//	          "ResourceRecordSets": [
//	            {
//	              "Name": "a",
//	              "Type": "A",
//	              "RData": "192.0.2.1",
//	              "TTL": 600
//	            },
//	            ...
//	          ]
//	        }
//	      },
//	      "SettingsHash": "ffffffffffffffffffffffffffffffff",
//	      "Status": {
//	        "Zone": "example.com",
//	        "NS": [
//	          "ns1.gslbN.sakura.ne.jp",
//	          "ns2.gslbN.sakura.ne.jp"
//	        ]
//	      },
//	      "ServiceClass": "cloud/dns",
//	      "Availability": "available",
//	      "CreatedAt": "2006-01-02T15:04:05+07:00",
//	      "ModifiedAt": "2006-01-02T15:04:05+07:00",
//	      "Provider": {
//	        "ID": 9999999,
//	        "Class": "dns",
//	        "Name": "gslbN.sakura.ne.jp",
//	        "ServiceClass": "cloud/dns"
//	      },
//	      "Icon": null,
//	      "Tags": []
//	    }
//	  ],
//	  "is_ok": true
//	}
//
// References:
//
//   - https://manual.sakura.ad.jp/cloud-api/1.1/appliance/#get_commonserviceitem
type responseCommonServiceItems struct {
	From               int                 `json:"From"`
	Count              int                 `json:"Count"`
	Total              int                 `json:"Total"`
	CommonServiceItems []commonServiceItem `json:"CommonServiceItems"`
	IsOk               bool                `json:"is_ok"`
}

// responseCommonServiceItem is the body structure of the success response to get a zone or update zone data.
//
// Request:
//
//	GET /commonserviceitem/:commonserviceitemid
//	PUT /commonserviceitem/:commonserviceitemid
//
// Response body structure:
//
//	{
//	  "CommonServiceItem": {
//	    "ID": "999999999999",
//	    "Name": "example.com",
//	    "Description": "",
//	    "Settings": {
//	      "DNS": {
//	        "ResourceRecordSets": [
//	          {
//	            "Name": "a",
//	            "Type": "A",
//	            "RData": "192.0.2.1",
//	            "TTL": 600
//	          },
//	          ...
//	        ]
//	      }
//	    },
//	    "SettingsHash": "ffffffffffffffffffffffffffffffff",
//	    "Status": {
//	      "Zone": "example.com",
//	      "NS": [
//	        "ns1.gslbN.sakura.ne.jp",
//	        "ns2.gslbN.sakura.ne.jp"
//	      ]
//	    },
//	    "ServiceClass": "cloud/dns",
//	    "Availability": "available",
//	    "CreatedAt": "2006-01-02T15:04:05+07:00",
//	    "ModifiedAt": "2006-01-02T15:04:05+07:00",
//	    "Provider": {
//	      "ID": 9999999,
//	      "Class": "dns",
//	      "Name": "gslbN.sakura.ne.jp",
//	      "ServiceClass": "cloud/dns"
//	    },
//	    "Icon": null,
//	    "Tags": []
//	  },
//	  "Success": true,
//	  "is_ok": true
//	}
//
// References:
//
//   - https://manual.sakura.ad.jp/cloud-api/1.1/appliance/#get_commonserviceitem_commonserviceitemid
//   - https://manual.sakura.ad.jp/cloud-api/1.1/appliance/#put_commonserviceitem_commonserviceitemid
type responseCommonServiceItem struct {
	CommonServiceItem commonServiceItem `json:"CommonServiceItem"`
	Success           bool              `json:"Success"`
	IsOk              bool              `json:"is_ok"`
}

// errorResponse is the body structure of an error response.
//
// Response body structure:
//
//	{
//	  "is_fatal": true,
//	  "serial": "ffffffffffffffffffffffffffffffff",
//	  "status": "401 Unauthorized",
//	  "error_code": "unauthorized",
//	  "error_msg": "error-unauthorized"
//	}
type errorResponse struct {
	IsFatal   bool   `json:"is_fatal"`
	Serial    string `json:"serial"`
	Status    string `json:"status"`
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

// commonServiceItem is a resource structure.
type commonServiceItem struct {
	ID           string   `json:"ID,omitempty"`
	Name         string   `json:"Name,omitempty"`
	Settings     settings `json:"Settings"`
	Status       status   `json:"Status,omitempty"`
	ServiceClass string   `json:"ServiceClass,omitempty"`
	Provider     provider `json:"Provider,omitempty"`
}

// settings is a resource setting.
type settings struct {
	DNS dNS `json:"DNS"`
}

// dNS is a set of dNS resources.
type dNS struct {
	ResourceRecordSets []domainRecord `json:"ResourceRecordSets"`
}

// domainRecord is a resource record.
type domainRecord struct {
	Name  string `json:"Name"`
	Type  string `json:"Type"`
	RData string `json:"RData"`
	TTL   uint32 `json:"TTL,omitempty"`
}

// status is the metadata of a zone.
type status struct {
	Zone string   `json:"Zone,omitempty"`
	NS   []string `json:"NS,omitempty"`
}

// provider is the metadata of a service.
type provider struct {
	ID           int    `json:"ID,omitempty"`
	Class        string `json:"Class"`
	Name         string `json:"Name,omitempty"`
	ServiceClass string `json:"ServiceClass,omitempty"`
}

// sakuracloudAPI has information about the API of the Sakura Cloud.
type sakuracloudAPI struct {
	accessToken          string
	accessTokenSecret    string
	baseURL              url.URL
	httpClient           *http.Client
	commonServiceItemMap map[string]*commonServiceItem
}

func NewSakuracloudAPI(accessToken, accessTokenSecret, endpoint string) (*sakuracloudAPI, error) {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("endpoint_url parse error: %w", err)
	}

	return &sakuracloudAPI{
		accessToken:       accessToken,
		accessTokenSecret: accessTokenSecret,
		baseURL:           *baseURL,
		httpClient: &http.Client{
			Timeout: time.Minute,
		},
	}, nil
}

func (api *sakuracloudAPI) request(method, path string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(method, path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(api.accessToken, api.accessTokenSecret)
	req.Header.Add("Content-Type", "applicaiton/json; charset=UTF-8")
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		err := json.Unmarshal(respBody, &errResp)
		if err != nil {
			return nil, err
		}
		// Since an error_msg uses HTML entities, unescape it.
		return nil, fmt.Errorf("request failed: status: %s, serial: %s, error_code: %s, error_msg: %s", errResp.Status, errResp.Serial, errResp.ErrorCode, html.UnescapeString(errResp.ErrorMsg))
	}

	return respBody, nil
}

// getCommonServiceItems return all the zones in the account
func (api *sakuracloudAPI) getCommonServiceItems() ([]*commonServiceItem, error) {
	var items []*commonServiceItem

	nextFrom := 0
	count := 100
	for {
		u := api.baseURL.JoinPath("/commonserviceitem")

		if nextFrom > 0 {
			// The query string is similar to the flow-style YAML.
			//	{From: 0, Count: 10}
			query := fmt.Sprintf("{From: %d, Count: %d}", nextFrom, count)
			u.RawQuery = url.QueryEscape(query)
		}

		respBody, err := api.request(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		var respData responseCommonServiceItems
		err = json.Unmarshal(respBody, &respData)
		if err != nil {
			return nil, err
		}

		if items == nil {
			items = make([]*commonServiceItem, 0, respData.Total)
		}

		for _, item := range respData.CommonServiceItems {
			items = append(items, &item)
		}

		count = respData.Count
		nextFrom = respData.From + respData.Count
		if nextFrom == respData.Total {
			break
		}
	}

	return items, nil
}

// GetCommonServiceItemMap return all the zones in the account
func (api *sakuracloudAPI) GetCommonServiceItemMap() (map[string]*commonServiceItem, error) {
	if api.commonServiceItemMap != nil {
		return api.commonServiceItemMap, nil
	}

	items, err := api.getCommonServiceItems()
	if err != nil {
		return nil, err
	}

	api.commonServiceItemMap = make(map[string]*commonServiceItem, len(items))
	for _, item := range items {
		if item.ServiceClass != "cloud/dns" {
			continue
		}
		api.commonServiceItemMap[item.Status.Zone] = item
	}

	return api.commonServiceItemMap, nil
}

// postCommonServiceItem submits a CommonServiceItem to the API and create the zone.
func (api *sakuracloudAPI) postCommonServiceItem(reqItem requestCommonServiceItem) (*commonServiceItem, error) {
	reqBody, err := json.Marshal(reqItem)
	if err != nil {
		return nil, err
	}

	u := api.baseURL.JoinPath("/commonserviceitem")
	respBody, err := api.request(http.MethodPost, u.String(), reqBody)
	if err != nil {
		return nil, err
	}

	var respData responseCommonServiceItem
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData.CommonServiceItem, nil
}

// CreateZone submits a CommonServiceItem to the API and create the zone.
func (api *sakuracloudAPI) CreateZone(domain string) error {
	reqItem := requestCommonServiceItem{
		CommonServiceItem: commonServiceItem{
			Name: domain,
			Status: status{
				Zone: domain,
			},
			Settings: settings{
				DNS: dNS{
					ResourceRecordSets: []domainRecord{},
				},
			},
			Provider: provider{
				Class: "dns",
			},
		},
	}

	item, err := api.postCommonServiceItem(reqItem)
	if err != nil {
		return err
	}

	api.commonServiceItemMap[domain] = item
	return nil
}

// putCommonServiceItem submits a CommonServiceItem to the API and updates the zone data.
func (api *sakuracloudAPI) putCommonServiceItem(id string, reqItem requestCommonServiceItem) (*commonServiceItem, error) {
	reqBody, err := json.Marshal(reqItem)
	if err != nil {
		return nil, err
	}

	u := api.baseURL.JoinPath("/commonserviceitem/").JoinPath(id)
	respBody, err := api.request(http.MethodPut, u.String(), reqBody)
	if err != nil {
		return nil, err
	}

	var respData responseCommonServiceItem
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData.CommonServiceItem, nil
}

// UpdateZone submits a CommonServiceItem to the API and updates the zone data.
func (api *sakuracloudAPI) UpdateZone(domain string, domainRecords []domainRecord) error {
	drs := make([]domainRecord, 0, len(domainRecords)-2) // Removes 2 NS records.
	for _, r := range domainRecords {
		if r.Type == "NS" && r.Name == "@" {
			continue
		}
		drs = append(drs, r)
	}

	reqItem := requestCommonServiceItem{
		CommonServiceItem: commonServiceItem{
			Settings: settings{
				DNS: dNS{
					ResourceRecordSets: drs,
				},
			},
		},
	}

	item, err := api.putCommonServiceItem(api.commonServiceItemMap[domain].ID, reqItem)
	if err != nil {
		return err
	}

	api.commonServiceItemMap[domain] = item
	return nil
}
