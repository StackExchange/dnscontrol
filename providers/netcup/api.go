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

type api struct {
	domainIndex      map[string]string
	nameserversNames []string
	credentials      struct {
		apikey         string
		customernumber string
		sessionId      string
	}
}

func (api *api) createRecord(domain string, rec *record) error {
	rec.Delete = false
	data := paramUpdateRecords{
		Key:            api.credentials.apikey,
		SessionId:      api.credentials.sessionId,
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

func (api *api) deleteRecord(domain string, rec *record) error {
	rec.Delete = true
	data := paramUpdateRecords{
		Key:            api.credentials.apikey,
		SessionId:      api.credentials.sessionId,
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

func (api *api) modifyRecord(domain string, rec *record) error {
	rec.Delete = false
	data := paramUpdateRecords{
		Key:            api.credentials.apikey,
		SessionId:      api.credentials.sessionId,
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

func (api *api) getRecords(domain string) ([]record, error) {
	data := paramGetRecords{
		Key:            api.credentials.apikey,
		SessionId:      api.credentials.sessionId,
		CustomerNumber: api.credentials.customernumber,
		DomainName:     domain,
	}
	rawJson, err := api.get("infoDnsRecords", data)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to login to netcup: %s", err)
	}

	resp := &records{}
	json.Unmarshal(rawJson, &resp)
	return resp.Records, nil
}

func (api *api) login(apikey, password, customernumber string) error {
	data := paramLogin{
		Key:            apikey,
		Password:       password,
		CustomerNumber: customernumber,
	}
	rawJson, err := api.get("login", data)
	if err != nil {
		return fmt.Errorf("Error while trying to login to netcup: %s", err)
	}

	resp := &responseLogin{}
	json.Unmarshal(rawJson, &resp)
	api.credentials.apikey = apikey
	api.credentials.customernumber = customernumber
	api.credentials.sessionId = resp.SessionId
	return nil
}

func (api *api) logout() error {
	data := paramLogout{
		Key:            api.credentials.apikey,
		SessionId:      api.credentials.sessionId,
		CustomerNumber: api.credentials.customernumber,
	}
	_, err := api.get("logout", data)
	if err != nil {
		return fmt.Errorf("Error while trying to logout from netcup: %s", err)
	}
	api.credentials.apikey, api.credentials.sessionId, api.credentials.customernumber = "", "", ""
	return nil
}

func (api *api) get(action string, params interface{}) (json.RawMessage, error) {
	reqParam := request{
		Action: action,
		Param:  params,
	}
	reqJson, _ := json.Marshal(reqParam)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqJson))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)

	respData := &response{}
	err = json.Unmarshal(bodyString, &respData)

	// Yeah, netcup implemented an empty recordset as an error - don't ask.
	if action == "infoDnsRecords" && respData.StatusCode == 5029 {
		emptyRecords, _ := json.Marshal(records{})
		return emptyRecords, nil
	}

	// Check for any errors and log them
	if respData.StatusCode != 2000 && (action == "") {
		return nil, fmt.Errorf("Netcup API error: %v\n%v\n", reqParam, respData)
	}

	return respData.Data, nil
}
