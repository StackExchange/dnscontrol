package realtimeregister

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type realtimeregisterApi struct {
	apikey   string
	endpoint string
}

type Zone struct {
	Records []Record `json:"records"`
}

type Record struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	Priority int    `json:"prio,omitempty"`
	TTL      int    `json:"ttl"`
}

const (
	endpoint        = "https://api.yoursrs.com/v2/domains/%s/zone"
	endpointSandbox = "http://localhost:8080/srs/services/domains/%s/zone"
)

func (api *realtimeregisterApi) get(domain string) (*Zone, error) {
	client := &http.Client{}
	url := fmt.Sprintf(api.endpoint, domain)
	req, _ := http.NewRequest(
		"GET",
		url,
		nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "ApiKey "+api.apikey)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	bodyString, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("realtime Register API error on get request: %d, %s, %s", resp.StatusCode,
			url, string(bodyString))
	}

	respData := &Zone{}

	err = json.Unmarshal(bodyString, &respData)
	if err != nil {
		return respData, err
	}

	return respData, nil
}

func (api *realtimeregisterApi) post(domain string, body *Zone) error {
	client := &http.Client{}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, string(bodyBytes)+"\n")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(api.endpoint, domain)+"/update",
		bytes.NewReader(bodyBytes),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "ApiKey "+api.apikey)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	bodyString, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("realtime register API error: %s", bodyString)
	}
	return nil
}
