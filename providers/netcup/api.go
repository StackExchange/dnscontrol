package netcup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	endpoint = "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON"
)

type netcupProvider struct {
	domainIndex      map[string]string
	nameserversNames []string
	credentials      struct {
		apikey         string
		customernumber string
		sessionID      string
	}
}

func (api *netcupProvider) createRecord(domain string, rec *record) error {
	rec.Delete = false
	data := paramUpdateRecords{
		Key:            api.credentials.apikey,
		SessionID:      api.credentials.sessionID,
		CustomerNumber: api.credentials.customernumber,
		DomainName:     domain,
		RecordSet: records{Records: []record{
			*rec,
		}},
	}
	_, err := api.get("updateDnsRecords", data)
	if err != nil {
		return fmt.Errorf("error while trying to create a record: %s", err)
	}
	return nil
}

func (api *netcupProvider) deleteRecord(domain string, rec *record) error {
	rec.Delete = true
	data := paramUpdateRecords{
		Key:            api.credentials.apikey,
		SessionID:      api.credentials.sessionID,
		CustomerNumber: api.credentials.customernumber,
		DomainName:     domain,
		RecordSet: records{Records: []record{
			*rec,
		}},
	}
	_, err := api.get("updateDnsRecords", data)
	if err != nil {
		return fmt.Errorf("error while trying to delete a record: %s", err)
	}
	return nil
}

func (api *netcupProvider) modifyRecord(domain string, rec *record) error {
	rec.Delete = false
	data := paramUpdateRecords{
		Key:            api.credentials.apikey,
		SessionID:      api.credentials.sessionID,
		CustomerNumber: api.credentials.customernumber,
		DomainName:     domain,
		RecordSet: records{Records: []record{
			*rec,
		}},
	}
	_, err := api.get("updateDnsRecords", data)
	if err != nil {
		return fmt.Errorf("error while trying to modify a record: %s", err)
	}
	return nil
}

func (api *netcupProvider) getRecords(domain string) ([]record, error) {
	data := paramGetRecords{
		Key:            api.credentials.apikey,
		SessionID:      api.credentials.sessionID,
		CustomerNumber: api.credentials.customernumber,
		DomainName:     domain,
	}
	rawJSON, err := api.get("infoDnsRecords", data)
	if err != nil {
		return nil, fmt.Errorf("failed while trying to login (netcup): %s", err)
	}

	resp := &records{}
	json.Unmarshal(rawJSON, &resp)
	return resp.Records, nil
}

func (api *netcupProvider) login(apikey, password, customernumber string) error {
	data := paramLogin{
		Key:            apikey,
		Password:       password,
		CustomerNumber: customernumber,
	}
	rawJSON, err := api.get("login", data)
	if err != nil {
		return fmt.Errorf("failed while trying to login to (netcup): %s", err)
	}

	resp := &responseLogin{}
	json.Unmarshal(rawJSON, &resp)
	api.credentials.apikey = apikey
	api.credentials.customernumber = customernumber
	api.credentials.sessionID = resp.SessionID
	return nil
}

func (api *netcupProvider) logout() error {
	data := paramLogout{
		Key:            api.credentials.apikey,
		SessionID:      api.credentials.sessionID,
		CustomerNumber: api.credentials.customernumber,
	}
	_, err := api.get("logout", data)
	if err != nil {
		return fmt.Errorf("failed to logout from netcup: %s", err)
	}
	api.credentials.apikey, api.credentials.sessionID, api.credentials.customernumber = "", "", ""
	return nil
}

func (api *netcupProvider) get(action string, params interface{}) (json.RawMessage, error) {
	reqParam := request{
		Action: action,
		Param:  params,
	}
	reqJSON, _ := json.Marshal(reqParam)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqJSON))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)

	respData := &response{}
	err = json.Unmarshal(bodyString, &respData)
	if err != nil {
		return nil, err
	}

	// Yeah, netcup implemented an empty recordset as an error - don't ask.
	if action == "infoDnsRecords" && respData.StatusCode == 5029 {
		emptyRecords, _ := json.Marshal(records{})
		return emptyRecords, nil
	}

	// Check for any errors and log them
	if respData.StatusCode != 2000 && (action == "") {
		return nil, fmt.Errorf("netcup API error: %v\n%v", reqParam, respData)
	}

	return respData.Data, nil
}
