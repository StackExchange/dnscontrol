package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
)

const (
	baseURL           = "https://api.cloudflare.com/client/v4/"
	zonesURL          = baseURL + "zones/"
	recordsURL        = zonesURL + "%s/dns_records/"
	pageRulesURL      = zonesURL + "%s/pagerules/"
	singlePageRuleURL = pageRulesURL + "%s"
	singleRecordURL   = recordsURL + "%s"
)

// get list of domains for account. Cache so the ids can be looked up from domain name
func (c *cloudflareProvider) fetchDomainList() error {
	c.domainIndex = map[string]string{}
	c.nameservers = map[string][]string{}
	page := 1
	for {
		zr := &zoneResponse{}
		url := fmt.Sprintf("%s?page=%d&per_page=50", zonesURL, page)
		if err := c.get(url, zr); err != nil {
			return fmt.Errorf("failed fetching domain list from cloudflare: %s", err)
		}
		if !zr.Success {
			return fmt.Errorf("failed fetching domain list from cloudflare: %s", stringifyErrors(zr.Errors))
		}
		for _, zone := range zr.Result {
			c.domainIndex[zone.Name] = zone.ID
			c.nameservers[zone.Name] = append(c.nameservers[zone.Name], zone.Nameservers...)
		}
		ri := zr.ResultInfo
		if len(zr.Result) == 0 || ri.Page*ri.PerPage >= ri.TotalCount {
			break
		}
		page++
	}
	return nil
}

// get all records for a domain
func (c *cloudflareProvider) getRecordsForDomain(id string, domain string) ([]*models.RecordConfig, error) {
	url := fmt.Sprintf(recordsURL, id)
	page := 1
	records := []*models.RecordConfig{}
	for {
		reqURL := fmt.Sprintf("%s?page=%d&per_page=100", url, page)
		var data recordsResponse
		if err := c.get(reqURL, &data); err != nil {
			return nil, fmt.Errorf("failed fetching record list from cloudflare: %s", err)
		}
		if !data.Success {
			return nil, fmt.Errorf("failed fetching record list cloudflare: %s", stringifyErrors(data.Errors))
		}
		for _, rec := range data.Result {
			rt, err := rec.nativeToRecord(domain)
			if err != nil {
				return nil, err
			}
			records = append(records, rt)
		}
		ri := data.ResultInfo
		if len(data.Result) == 0 || ri.Page*ri.PerPage >= ri.TotalCount {
			break
		}
		page++
	}
	return records, nil
}

// create a correction to delete a record
func (c *cloudflareProvider) deleteRec(rec *cfRecord, domainID string) *models.Correction {
	return &models.Correction{
		Msg: fmt.Sprintf("DELETE record: %s %s %d %s (id=%s)", rec.Name, rec.Type, rec.TTL, rec.Content, rec.ID),
		F: func() error {
			endpoint := fmt.Sprintf(singleRecordURL, domainID, rec.ID)
			req, err := http.NewRequest("DELETE", endpoint, nil)
			if err != nil {
				return err
			}
			c.setHeaders(req)
			_, err = handleActionResponse(http.DefaultClient.Do(req))
			return err
		},
	}
}

func (c *cloudflareProvider) createZone(domainName string) (string, error) {
	type createZone struct {
		Name string `json:"name"`

		Account struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
	}
	var id string
	cz := &createZone{
		Name: domainName}

	if c.AccountID != "" || c.AccountName != "" {
		cz.Account.ID = c.AccountID
		cz.Account.Name = c.AccountName
	}

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(cz); err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", zonesURL, buf)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)
	id, err = handleActionResponse(http.DefaultClient.Do(req))
	return id, err
}

func cfDSData(rec *models.RecordConfig) *cfRecData {
	return &cfRecData{
		KeyTag:     rec.DsKeyTag,
		Algorithm:  rec.DsAlgorithm,
		DigestType: rec.DsDigestType,
		Digest:     rec.DsDigest,
	}
}

func cfSrvData(rec *models.RecordConfig) *cfRecData {
	serverParts := strings.Split(rec.GetLabelFQDN(), ".")
	c := &cfRecData{
		Service:  serverParts[0],
		Proto:    serverParts[1],
		Name:     strings.Join(serverParts[2:], "."),
		Port:     rec.SrvPort,
		Priority: rec.SrvPriority,
		Weight:   rec.SrvWeight,
	}
	c.Target = cfTarget(rec.GetTargetField())
	return c
}

func cfCaaData(rec *models.RecordConfig) *cfRecData {
	return &cfRecData{
		Tag:   rec.CaaTag,
		Flags: rec.CaaFlag,
		Value: rec.GetTargetField(),
	}
}

func cfTlsaData(rec *models.RecordConfig) *cfRecData {
	return &cfRecData{
		Usage:        rec.TlsaUsage,
		Selector:     rec.TlsaSelector,
		MatchingType: rec.TlsaMatchingType,
		Certificate:  rec.GetTargetField(),
	}
}

func cfSshfpData(rec *models.RecordConfig) *cfRecData {
	return &cfRecData{
		Algorithm:   rec.SshfpAlgorithm,
		HashType:    rec.SshfpFingerprint,
		Fingerprint: rec.GetTargetField(),
	}
}

func (c *cloudflareProvider) createRec(rec *models.RecordConfig, domainID string) []*models.Correction {
	type createRecord struct {
		Name     string     `json:"name"`
		Type     string     `json:"type"`
		Content  string     `json:"content"`
		TTL      uint32     `json:"ttl"`
		Priority uint16     `json:"priority"`
		Data     *cfRecData `json:"data"`
	}
	var id string
	content := rec.GetTargetField()
	if rec.Metadata[metaOriginalIP] != "" {
		content = rec.Metadata[metaOriginalIP]
	}
	prio := ""
	if rec.Type == "MX" {
		prio = fmt.Sprintf(" %d ", rec.MxPreference)
	}
	if rec.Type == "DS" {
		content = fmt.Sprintf("%d %d %d %s", rec.DsKeyTag, rec.DsAlgorithm, rec.DsDigestType, rec.DsDigest)
	}
	arr := []*models.Correction{{
		Msg: fmt.Sprintf("CREATE record: %s %s %d%s %s", rec.GetLabel(), rec.Type, rec.TTL, prio, content),
		F: func() error {

			cf := &createRecord{
				Name:     rec.GetLabel(),
				Type:     rec.Type,
				TTL:      rec.TTL,
				Content:  content,
				Priority: rec.MxPreference,
			}
			if rec.Type == "SRV" {
				cf.Data = cfSrvData(rec)
				cf.Name = rec.GetLabelFQDN()
			} else if rec.Type == "CAA" {
				cf.Data = cfCaaData(rec)
				cf.Name = rec.GetLabelFQDN()
				cf.Content = ""
			} else if rec.Type == "TLSA" {
				cf.Data = cfTlsaData(rec)
				cf.Name = rec.GetLabelFQDN()
			} else if rec.Type == "SSHFP" {
				cf.Data = cfSshfpData(rec)
				cf.Name = rec.GetLabelFQDN()
			} else if rec.Type == "DS" {
				cf.Data = cfDSData(rec)
			}
			endpoint := fmt.Sprintf(recordsURL, domainID)
			buf := &bytes.Buffer{}
			encoder := json.NewEncoder(buf)
			if err := encoder.Encode(cf); err != nil {
				return err
			}
			req, err := http.NewRequest("POST", endpoint, buf)
			if err != nil {
				return err
			}
			c.setHeaders(req)
			id, err = handleActionResponse(http.DefaultClient.Do(req))
			return err
		},
	}}
	if rec.Metadata[metaProxy] != "off" {
		arr = append(arr, &models.Correction{
			Msg: fmt.Sprintf("ACTIVATE PROXY for new record %s %s %d %s", rec.GetLabel(), rec.Type, rec.TTL, rec.GetTargetField()),
			F:   func() error { return c.modifyRecord(domainID, id, true, rec) },
		})
	}
	return arr
}

func (c *cloudflareProvider) modifyRecord(domainID, recID string, proxied bool, rec *models.RecordConfig) error {
	if domainID == "" || recID == "" {
		return fmt.Errorf("cannot modify record if domain or record id are empty")
	}
	type record struct {
		ID       string     `json:"id"`
		Proxied  bool       `json:"proxied"`
		Name     string     `json:"name"`
		Type     string     `json:"type"`
		Content  string     `json:"content"`
		Priority uint16     `json:"priority"`
		TTL      uint32     `json:"ttl"`
		Data     *cfRecData `json:"data"`
	}
	r := record{
		ID:       recID,
		Proxied:  proxied,
		Name:     rec.GetLabel(),
		Type:     rec.Type,
		Content:  rec.GetTargetField(),
		Priority: rec.MxPreference,
		TTL:      rec.TTL,
		Data:     nil,
	}
	if rec.Type == "SRV" {
		r.Data = cfSrvData(rec)
		r.Name = rec.GetLabelFQDN()
	} else if rec.Type == "CAA" {
		r.Data = cfCaaData(rec)
		r.Name = rec.GetLabelFQDN()
		r.Content = ""
	} else if rec.Type == "TLSA" {
		r.Data = cfTlsaData(rec)
		r.Name = rec.GetLabelFQDN()
	} else if rec.Type == "SSHFP" {
		r.Data = cfSshfpData(rec)
		r.Name = rec.GetLabelFQDN()
	} else if rec.Type == "DS" {
		r.Data = cfDSData(rec)
		r.Content = ""
	}
	endpoint := fmt.Sprintf(singleRecordURL, domainID, recID)
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(r); err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", endpoint, buf)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	_, err = handleActionResponse(http.DefaultClient.Do(req))
	return err
}

// change universal ssl state
func (c *cloudflareProvider) changeUniversalSSL(domainID string, state bool) error {
	type setUniversalSSL struct {
		Enabled bool `json:"enabled"`
	}
	us := &setUniversalSSL{
		Enabled: state,
	}

	// create json
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(us); err != nil {
		return err
	}

	// send request.
	endpoint := fmt.Sprintf(zonesURL+"%s/ssl/universal/settings", domainID)
	req, err := http.NewRequest("PATCH", endpoint, buf)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	_, err = handleActionResponse(http.DefaultClient.Do(req))

	return err
}

// change universal ssl state
func (c *cloudflareProvider) getUniversalSSL(domainID string) (bool, error) {
	type universalSSLResponse struct {
		Success  bool          `json:"success"`
		Errors   []interface{} `json:"errors"`
		Messages []interface{} `json:"messages"`
		Result   struct {
			Enabled bool `json:"enabled"`
		} `json:"result"`
	}

	// send request.
	endpoint := fmt.Sprintf(zonesURL+"%s/ssl/universal/settings", domainID)
	var result universalSSLResponse
	err := c.get(endpoint, &result)
	if err != nil {
		return true, err
	}

	return result.Result.Enabled, err
}

// common error handling for all action responses
func handleActionResponse(resp *http.Response, err error) (id string, e error) {
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result := &basicResponse{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(result); err != nil {
		return "", fmt.Errorf("unknown error. Status code: %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf(stringifyErrors(result.Errors))
	}
	return result.Result.ID, nil
}

func (c *cloudflareProvider) setHeaders(req *http.Request) {
	if len(c.APIToken) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	} else {
		req.Header.Set("X-Auth-Key", c.APIKey)
		req.Header.Set("X-Auth-Email", c.APIUser)
	}
}

// generic get handler. makes request and unmarshalls response to given interface
func (c *cloudflareProvider) get(endpoint string, target interface{}) error {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		dat, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(dat))
		return fmt.Errorf("bad status code from cloudflare: %d not 200", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(target)
}

func (c *cloudflareProvider) getPageRules(id string, domain string) ([]*models.RecordConfig, error) {
	url := fmt.Sprintf(pageRulesURL, id)
	data := pageRuleResponse{}
	if err := c.get(url, &data); err != nil {
		return nil, fmt.Errorf("failed fetching page rule list from cloudflare: %s", err)
	}
	if !data.Success {
		return nil, fmt.Errorf("failed fetching page rule list cloudflare: %s", stringifyErrors(data.Errors))
	}
	recs := []*models.RecordConfig{}
	for _, pr := range data.Result {
		// only interested in forwarding rules. Lets be very specific, and skip anything else
		if len(pr.Actions) != 1 || len(pr.Targets) != 1 {
			continue
		}
		if pr.Actions[0].ID != "forwarding_url" {
			continue
		}
		err := json.Unmarshal([]byte(pr.Actions[0].Value), &pr.ForwardingInfo)
		if err != nil {
			return nil, err
		}
		var thisPr = pr
		r := &models.RecordConfig{
			Type:     "PAGE_RULE",
			Original: thisPr,
			TTL:      1,
		}
		r.SetLabel("@", domain)
		r.SetTarget(fmt.Sprintf("%s,%s,%d,%d", // $FROM,$TO,$PRIO,$CODE
			pr.Targets[0].Constraint.Value,
			pr.ForwardingInfo.URL,
			pr.Priority,
			pr.ForwardingInfo.StatusCode))
		recs = append(recs, r)
	}
	return recs, nil
}

func (c *cloudflareProvider) deletePageRule(recordID, domainID string) error {
	endpoint := fmt.Sprintf(singlePageRuleURL, domainID, recordID)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	_, err = handleActionResponse(http.DefaultClient.Do(req))
	return err
}

func (c *cloudflareProvider) updatePageRule(recordID, domainID string, target string) error {
	if err := c.deletePageRule(recordID, domainID); err != nil {
		return err
	}
	return c.createPageRule(domainID, target)
}

func (c *cloudflareProvider) createPageRule(domainID string, target string) error {
	endpoint := fmt.Sprintf(pageRulesURL, domainID)
	return c.sendPageRule(endpoint, "POST", target)
}

func (c *cloudflareProvider) sendPageRule(endpoint, method string, data string) error {
	// from to priority code
	parts := strings.Split(data, ",")
	priority, _ := strconv.Atoi(parts[2])
	code, _ := strconv.Atoi(parts[3])
	fwdInfo := &pageRuleFwdInfo{
		StatusCode: code,
		URL:        parts[1],
	}
	dat, _ := json.Marshal(fwdInfo)
	pr := &pageRule{
		Status:   "active",
		Priority: priority,
		Targets: []pageRuleTarget{
			{Target: "url", Constraint: pageRuleConstraint{Operator: "matches", Value: parts[0]}},
		},
		Actions: []pageRuleAction{
			{ID: "forwarding_url", Value: json.RawMessage(dat)},
		},
	}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(pr); err != nil {
		return err
	}
	req, err := http.NewRequest(method, endpoint, buf)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	_, err = handleActionResponse(http.DefaultClient.Do(req))
	return err
}

func stringifyErrors(errors []interface{}) string {
	dat, err := json.Marshal(errors)
	if err != nil {
		return "???"
	}
	return string(dat)
}

type recordsResponse struct {
	basicResponse
	Result     []*cfRecord `json:"result"`
	ResultInfo pagingInfo  `json:"result_info"`
}

type basicResponse struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   struct {
		ID string `json:"id"`
	} `json:"result"`
}

type pageRuleResponse struct {
	basicResponse
	Result     []*pageRule `json:"result"`
	ResultInfo pagingInfo  `json:"result_info"`
}

type pageRule struct {
	ID             string           `json:"id,omitempty"`
	Targets        []pageRuleTarget `json:"targets"`
	Actions        []pageRuleAction `json:"actions"`
	Priority       int              `json:"priority"`
	Status         string           `json:"status"`
	ModifiedOn     time.Time        `json:"modified_on,omitempty"`
	CreatedOn      time.Time        `json:"created_on,omitempty"`
	ForwardingInfo *pageRuleFwdInfo `json:"-"`
}

type pageRuleTarget struct {
	Target     string             `json:"target"`
	Constraint pageRuleConstraint `json:"constraint"`
}

type pageRuleConstraint struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type pageRuleAction struct {
	ID    string          `json:"id"`
	Value json.RawMessage `json:"value"`
}

type pageRuleFwdInfo struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}

type zoneResponse struct {
	basicResponse
	Result []struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Nameservers []string `json:"name_servers"`
	} `json:"result"`
	ResultInfo pagingInfo `json:"result_info"`
}

type pagingInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}
