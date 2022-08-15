package cloudflare

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/cloudflare/cloudflare-go"
)

// get list of domains for account. Cache so the ids can be looked up from domain name
func (c *cloudflareProvider) fetchDomainList() error {
	c.domainIndex = map[string]string{}
	c.nameservers = map[string][]string{}
	zones, err := c.cfClient.ListZones(context.Background())
	if err != nil {
		return fmt.Errorf("failed fetching domain list from cloudflare(%q): %s", c.cfClient.APIEmail, err)
	}

	for _, zone := range zones {
		c.domainIndex[zone.Name] = zone.ID
		c.nameservers[zone.Name] = append(c.nameservers[zone.Name], zone.NameServers...)
	}

	return nil
}

// get all records for a domain
func (c *cloudflareProvider) getRecordsForDomain(id string, domain string) ([]*models.RecordConfig, error) {
	records := []*models.RecordConfig{}
	rrs, err := c.cfClient.DNSRecords(context.Background(), id, cloudflare.DNSRecord{})
	if err != nil {
		return nil, fmt.Errorf("failed fetching record list from cloudflare(%q): %w", c.cfClient.APIEmail, err)
	}
	for _, rec := range rrs {
		rt, err := c.nativeToRecord(domain, rec)
		if err != nil {
			return nil, err
		}
		records = append(records, rt)
	}
	return records, nil
}

// create a correction to delete a record
func (c *cloudflareProvider) deleteRec(rec cloudflare.DNSRecord, domainID string) *models.Correction {
	return &models.Correction{
		Msg: fmt.Sprintf("DELETE record: %s %s %d %q (id=%s)", rec.Name, rec.Type, rec.TTL, rec.Content, rec.ID),
		F: func() error {
			err := c.cfClient.DeleteDNSRecord(context.Background(), domainID, rec.ID)
			return err
		},
	}
}

func (c *cloudflareProvider) createZone(domainName string) (string, error) {
	zone, err := c.cfClient.CreateZone(context.Background(), domainName, false, cloudflare.Account{ID: c.cfClient.AccountID}, "full")
	return zone.ID, err
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
	var id string
	content := rec.GetTargetField()
	if rec.Metadata[metaOriginalIP] != "" {
		content = rec.Metadata[metaOriginalIP]
	}
	prio := ""
	if rec.Type == "MX" {
		prio = fmt.Sprintf(" %d ", rec.MxPreference)
	}
	if rec.Type == "TXT" {
		content = rec.GetTargetTXTJoined()
	}
	if rec.Type == "DS" {
		content = fmt.Sprintf("%d %d %d %s", rec.DsKeyTag, rec.DsAlgorithm, rec.DsDigestType, rec.DsDigest)
	}
	arr := []*models.Correction{{
		Msg: fmt.Sprintf("CREATE record: %s %s %d%s %s", rec.GetLabel(), rec.Type, rec.TTL, prio, content),
		F: func() error {
			cf := cloudflare.DNSRecord{
				Name:     rec.GetLabel(),
				Type:     rec.Type,
				TTL:      int(rec.TTL),
				Content:  content,
				Priority: &rec.MxPreference,
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
			resp, err := c.cfClient.CreateDNSRecord(context.Background(), domainID, cf)
			if err != nil {
				return err
			}
			// Updating id (from the outer scope) by side-effect, required for updating proxy mode
			id = resp.Result.ID
			return nil
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

	r := cloudflare.DNSRecord{
		ID:       recID,
		Proxied:  &proxied,
		Name:     rec.GetLabel(),
		Type:     rec.Type,
		Content:  rec.GetTargetField(),
		Priority: &rec.MxPreference,
		TTL:      int(rec.TTL),
	}
	if rec.Type == "TXT" {
		r.Content = rec.GetTargetTXTJoined()
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
	return c.cfClient.UpdateDNSRecord(context.Background(), domainID, recID, r)
}

// change universal ssl state
func (c *cloudflareProvider) changeUniversalSSL(domainID string, state bool) error {
	_, err := c.cfClient.EditUniversalSSLSetting(context.Background(), domainID, cloudflare.UniversalSSLSetting{Enabled: state})
	return err
}

// get universal ssl state
func (c *cloudflareProvider) getUniversalSSL(domainID string) (bool, error) {
	result, err := c.cfClient.UniversalSSLSettingDetails(context.Background(), domainID)
	return result.Enabled, err
}

func (c *cloudflareProvider) getPageRules(id string, domain string) ([]*models.RecordConfig, error) {
	rules, err := c.cfClient.ListPageRules(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("failed fetching page rule list cloudflare: %s", err)
	}
	recs := []*models.RecordConfig{}
	for _, pr := range rules {
		// only interested in forwarding rules. Lets be very specific, and skip anything else
		if len(pr.Actions) != 1 || len(pr.Targets) != 1 {
			continue
		}
		if pr.Actions[0].ID != "forwarding_url" {
			continue
		}
		value := pr.Actions[0].Value.(map[string]interface{})
		var thisPr = pr
		r := &models.RecordConfig{
			Type:     "PAGE_RULE",
			Original: thisPr,
			TTL:      1,
		}
		r.SetLabel("@", domain)
		r.SetTarget(fmt.Sprintf("%s,%s,%d,%d", // $FROM,$TO,$PRIO,$CODE
			pr.Targets[0].Constraint.Value,
			value["url"],
			pr.Priority,
			intZero(value["status_code"])))
		recs = append(recs, r)
	}
	return recs, nil
}

func (c *cloudflareProvider) deletePageRule(recordID, domainID string) error {
	return c.cfClient.DeletePageRule(context.Background(), domainID, recordID)
}

func (c *cloudflareProvider) updatePageRule(recordID, domainID string, target string) error {
	// maybe someday?
	//c.apiProvider.UpdatePageRule(context.Background(), domainId, recordID, )
	if err := c.deletePageRule(recordID, domainID); err != nil {
		return err
	}
	return c.createPageRule(domainID, target)
}

func (c *cloudflareProvider) createPageRule(domainID string, target string) error {
	// from to priority code
	parts := strings.Split(target, ",")
	priority, _ := strconv.Atoi(parts[2])
	code, _ := strconv.Atoi(parts[3])
	pr := cloudflare.PageRule{
		Status:   "active",
		Priority: priority,
		Targets: []cloudflare.PageRuleTarget{
			{Target: "url", Constraint: pageRuleConstraint{Operator: "matches", Value: parts[0]}},
		},
		Actions: []cloudflare.PageRuleAction{
			{ID: "forwarding_url", Value: &pageRuleFwdInfo{
				StatusCode: code,
				URL:        parts[1],
			}},
		},
	}
	_, err := c.cfClient.CreatePageRule(context.Background(), domainID, pr)
	return err
}

func (c *cloudflareProvider) getWorkerRoutes(id string, domain string) ([]*models.RecordConfig, error) {
	res, err := c.cfClient.ListWorkerRoutes(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("failed fetching worker route list cloudflare: %s", err)
	}

	recs := []*models.RecordConfig{}
	for _, pr := range res.Routes {
		var thisPr = pr
		r := &models.RecordConfig{
			Type:     "WORKER_ROUTE",
			Original: thisPr,
			TTL:      1,
		}
		r.SetLabel("@", domain)
		r.SetTarget(fmt.Sprintf("%s,%s", // $PATTERN,$SCRIPT
			pr.Pattern,
			pr.Script))
		recs = append(recs, r)
	}
	return recs, nil
}

func (c *cloudflareProvider) deleteWorkerRoute(recordID, domainID string) error {
	_, err := c.cfClient.DeleteWorkerRoute(context.Background(), domainID, recordID)
	return err
}

func (c *cloudflareProvider) updateWorkerRoute(recordID, domainID string, target string) error {
	// Causing stack overflow (!?)
	// return c.updateWorkerRoute(recordID, domainID, target)

	if err := c.deleteWorkerRoute(recordID, domainID); err != nil {
		return err
	}
	return c.createWorkerRoute(domainID, target)
}

func (c *cloudflareProvider) createWorkerRoute(domainID string, target string) error {
	// $PATTERN,$SCRIPT
	parts := strings.Split(target, ",")
	if len(parts) != 2 {
		return fmt.Errorf("unexpected target: '%s' (expected: 'PATTERN,SCRIPT')", target)
	}
	wr := cloudflare.WorkerRoute{
		Pattern: parts[0],
		Script:  parts[1],
	}

	_, err := c.cfClient.CreateWorkerRoute(context.Background(), domainID, wr)
	return err
}

func (c *cloudflareProvider) createTestWorker(workerName string) error {
	wrp := cloudflare.WorkerRequestParams{
		ZoneID:     "",
		ScriptName: workerName,
	}

	script := `
		addEventListener("fetch", (event) => {
			event.respondWith(
				new Response("Ok.", { status: 200 })
			);
	  	});`

	_, err := c.cfClient.UploadWorker(context.Background(), &wrp, script)
	return err
}

// https://github.com/dominikh/go-tools/issues/1137 which is a dup of
// https://github.com/dominikh/go-tools/issues/810
//
//lint:ignore U1000 false positive due to
type pageRuleConstraint struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type pageRuleFwdInfo struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}
