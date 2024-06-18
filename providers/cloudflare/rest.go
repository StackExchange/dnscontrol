package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/net/idna"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/cloudflare/cloudflare-go"
)

// get list of domains for account. Cache so the ids can be looked up from domain name
// The caller must do all locking.
func (c *cloudflareProvider) cacheDomainList() error {
	if c.domainIndex != nil {
		return nil
	}

	c.domainIndex = map[string]string{}
	c.nameservers = map[string][]string{}
	//fmt.Printf("DEBUG: CLOUDFLARE POPULATING CACHE\n")
	zones, err := c.cfClient.ListZones(context.Background())
	if err != nil {
		return fmt.Errorf("failed fetching domain list from cloudflare(%q): %s", c.cfClient.APIEmail, err)
	}

	for _, zone := range zones {
		if encoded, err := idna.ToASCII(zone.Name); err == nil && encoded != zone.Name {
			if _, ok := c.domainIndex[encoded]; ok {
				fmt.Printf("WARNING: Zone %q appears twice in this cloudflare account\n", encoded)
			}
			c.domainIndex[encoded] = zone.ID
			c.nameservers[encoded] = zone.NameServers
		}
		if _, ok := c.domainIndex[zone.Name]; ok {
			fmt.Printf("WARNING: Zone %q appears twice in this cloudflare account\n", zone.Name)
		}
		c.domainIndex[zone.Name] = zone.ID
		c.nameservers[zone.Name] = zone.NameServers
	}

	return nil
}

// get all records for a domain
func (c *cloudflareProvider) getRecordsForDomain(id string, domain string) ([]*models.RecordConfig, error) {
	records := []*models.RecordConfig{}
	rrs, _, err := c.cfClient.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(id), cloudflare.ListDNSRecordsParams{})
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

func (c *cloudflareProvider) deleteDNSRecord(rec cloudflare.DNSRecord, domainID string) error {
	return c.cfClient.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(domainID), rec.ID)
}

func (c *cloudflareProvider) createZone(domainName string) (string, error) {
	zone, err := c.cfClient.CreateZone(context.Background(), domainName, false, cloudflare.Account{ID: c.accountID}, "full")
	return zone.ID, err
}

func cfDnskeyData(rec *models.RecordConfig) *cfRecData {
	return &cfRecData{
		Algorithm: rec.DnskeyAlgorithm,
		Flags:     rec.DnskeyFlags,
		Protocol:  rec.DnskeyProtocol,
		PublicKey: rec.DnskeyPublicKey,
	}
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
		Flags: uint16(rec.CaaFlag),
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

func cfSvcbData(rec *models.RecordConfig) *cfRecData {
	return &cfRecData{
		Priority: rec.SvcPriority,
		Target:   cfTarget(rec.GetTargetField()),
		Value:    rec.SvcParams,
	}
}

func cfNaptrData(rec *models.RecordConfig) *cfNaptrRecData {
	return &cfNaptrRecData{
		Flags:       rec.NaptrFlags,
		Order:       rec.NaptrOrder,
		Preference:  rec.NaptrPreference,
		Regex:       rec.NaptrRegexp,
		Replacement: rec.GetTargetField(),
		Service:     rec.NaptrService,
	}
}

func (c *cloudflareProvider) createRecDiff2(rec *models.RecordConfig, domainID string, msg string) []*models.Correction {

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
	if msg == "" {
		msg = fmt.Sprintf("CREATE record: %s %s %d%s %s", rec.GetLabel(), rec.Type, rec.TTL, prio, content)
	}
	if rec.Metadata[metaProxy] == "on" || rec.Metadata[metaProxy] == "full" {
		msg = msg + fmt.Sprintf("\nACTIVATE PROXY for new record %s %s %d %s", rec.GetLabel(), rec.Type, rec.TTL, rec.GetTargetField())
	}
	arr := []*models.Correction{{
		Msg: msg,
		F: func() error {
			cf := cloudflare.CreateDNSRecordParams{
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
			} else if rec.Type == "DNSKEY" {
				cf.Data = cfDnskeyData(rec)
			} else if rec.Type == "DS" {
				cf.Data = cfDSData(rec)
			} else if rec.Type == "NAPTR" {
				cf.Data = cfNaptrData(rec)
				cf.Name = rec.GetLabelFQDN()
			} else if rec.Type == "HTTPS" || rec.Type == "SVCB" {
				cf.Data = cfSvcbData(rec)
			}
			resp, err := c.cfClient.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(domainID), cf)
			if err != nil {
				return err
			}
			// Records are created with the proxy off. If proxy should be
			// enabled, we do a second API call.
			resultID := resp.ID
			if rec.Metadata[metaProxy] == "on" || rec.Metadata[metaProxy] == "full" {
				return c.modifyRecord(domainID, resultID, true, rec)
			}
			return nil
		},
	}}
	return arr
}

func (c *cloudflareProvider) modifyRecord(domainID, recID string, proxied bool, rec *models.RecordConfig) error {
	if domainID == "" || recID == "" {
		return fmt.Errorf("cannot modify record if domain or record id are empty")
	}

	r := cloudflare.UpdateDNSRecordParams{
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
	} else if rec.Type == "DNSKEY" {
		r.Data = cfDnskeyData(rec)
		r.Content = ""
	} else if rec.Type == "DS" {
		r.Data = cfDSData(rec)
		r.Content = ""
	} else if rec.Type == "NAPTR" {
		r.Data = cfNaptrData(rec)
		r.Name = rec.GetLabelFQDN()
	} else if rec.Type == "HTTPS" || rec.Type == "SVCB" {
		r.Data = cfSvcbData(rec)
	}
	_, err := c.cfClient.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(domainID), r)
	return err
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

func (c *cloudflareProvider) getSingleRedirects(id string, domain string) ([]*models.RecordConfig, error) {
	rules, err := c.cfClient.GetEntrypointRuleset(context.Background(), cloudflare.ZoneIdentifier(id), "http_request_dynamic_redirect")
	if err != nil {
		return nil, fmt.Errorf("failed fetching redirect rule list cloudflare: %s", err)
	}
	//var rulelist []cloudflare.RulesetRule
	//rulelist = rules.Rules
	//rulelist := rules.Rules

	//printer.Printf("DEBUG: rules %+v\n", rules)
	recs := []*models.RecordConfig{}
	for _, pr := range rules.Rules {
		//printer.Printf("DEBUG: %+v\n", pr)

		var thisPr = pr
		r := &models.RecordConfig{
			Type:     "CLOUDFLAREAPI_SINGLE_REDIRECT",
			Original: thisPr,
			TTL:      1,
		}
		r.SetLabel("@", domain)

		// Extract the valuables from the rule, use it to make the sr:
		srMatcher := pr.Expression
		srReplacement := pr.ActionParameters.FromValue.TargetURL.Expression
		code := int(pr.ActionParameters.FromValue.StatusCode)
		sr := newCfsrFromAPIData(srMatcher, srReplacement, code)
		//sr.SRRRuleList = rulelist
		//printer.Printf("DEBUG: DESCRIPTION = %v\n", pr.Description)
		sr.SRDisplay = pr.Description
		// printer.Printf("DEBUG: PR = %+v\n", pr)
		// printer.Printf("DEBUG: rules = %+v\n", rules)
		sr.SRRRulesetID = rules.ID
		sr.SRRRulesetRuleID = pr.ID //correct

		r.CloudflareRedirect = sr
		r.SetTarget(pr.Description)

		recs = append(recs, r)
	}

	return recs, nil
}

func (c *cloudflareProvider) createSingleRedirect(domainID string, cfr models.CloudflareSingleRedirectConfig) error {

	//printer.Printf("DEBUG: createSingleRedir: d=%v crf=%+v\n", domainID, cfr)
	// Asumption for target:

	newSingleRedirectRulesActionParameters := cloudflare.RulesetRuleActionParameters{}
	newSingleRedirectRule := cloudflare.RulesetRule{}
	newSingleRedirectRules := []cloudflare.RulesetRule{}
	newSingleRedirectRules = append(newSingleRedirectRules, newSingleRedirectRule)
	newSingleRedirect := cloudflare.UpdateEntrypointRulesetParams{}

	// Preserve query string
	preserveQueryString := true
	newSingleRedirectRulesActionParameters.FromValue = &cloudflare.RulesetRuleActionParametersFromValue{}
	// Redirect status code
	newSingleRedirectRulesActionParameters.FromValue.StatusCode = uint16(cfr.Code)
	// Incoming request expression
	newSingleRedirectRules[0].Expression = cfr.SRMatcher
	// Redirect expression
	newSingleRedirectRulesActionParameters.FromValue.TargetURL.Expression = cfr.SRReplacement
	// Redirect name
	newSingleRedirectRules[0].Description = cfr.SRDisplay
	// Rule action, should always be redirect in this case
	newSingleRedirectRules[0].Action = "redirect"
	// Phase should always be http_request_dynamic_redirect
	newSingleRedirect.Phase = "http_request_dynamic_redirect"

	// Assigns the values in the nested structs
	newSingleRedirectRulesActionParameters.FromValue.PreserveQueryString = &preserveQueryString
	newSingleRedirectRules[0].ActionParameters = &newSingleRedirectRulesActionParameters

	// Get a list of current redirects so that the new redirect get appended to it
	rules, err := c.cfClient.GetEntrypointRuleset(context.Background(), cloudflare.ZoneIdentifier(domainID), "http_request_dynamic_redirect")
	if err != nil {
		return fmt.Errorf("failed fetching redirect rule list cloudflare: %s", err)
	}
	newSingleRedirect.Rules = newSingleRedirectRules
	newSingleRedirect.Rules = append(newSingleRedirect.Rules, rules.Rules...)

	_, err = c.cfClient.UpdateEntrypointRuleset(context.Background(), cloudflare.ZoneIdentifier(domainID), newSingleRedirect)

	return err
}

func (c *cloudflareProvider) deleteSingleRedirects(domainID string, cfr models.CloudflareSingleRedirectConfig) error {

	// This block should delete rules using the as is Cloudflare Golang lib in theory, need to debug why it isn't
	// updatedRuleset := cloudflare.UpdateEntrypointRulesetParams{}
	// updatedRulesetRules := []cloudflare.RulesetRule{}

	// rules, err := c.cfClient.GetEntrypointRuleset(context.Background(), cloudflare.ZoneIdentifier(domainID), "http_request_dynamic_redirect")
	// if err != nil {
	// 	return fmt.Errorf("failed fetching redirect rule list cloudflare: %s", err)
	// }

	// for _, rule := range rules.Rules {
	// 	if rule.ID != cfr.SRRRulesetRuleID {
	// 		updatedRulesetRules = append(updatedRulesetRules, rule)
	// 	} else {
	// 		printer.Printf("DEBUG: MATCH %v : %v\n", rule.ID, cfr.SRRRulesetRuleID)
	// 	}
	// }
	// updatedRuleset.Rules = updatedRulesetRules
	// _, err = c.cfClient.UpdateEntrypointRuleset(context.Background(), cloudflare.ZoneIdentifier(domainID), updatedRuleset)

	// Old Code

	// rules, err := c.cfClient.GetEntrypointRuleset(context.Background(), cloudflare.ZoneIdentifier(domainID), "http_request_dynamic_redirect")
	// if err != nil {
	// 	return err
	// }
	//printer.Printf("DEBUG: CALLING API DeleteRulesetRule: SRRRulesetID=%v, cfr.SRRRulesetRuleID=%v\n", cfr.SRRRulesetID, cfr.SRRRulesetRuleID)

	err := c.cfClient.DeleteRulesetRule(context.Background(), cloudflare.ZoneIdentifier(domainID), cfr.SRRRulesetID, cfr.SRRRulesetRuleID)
	// TODO(tlim): This is terrible.  It returns an error even when it is successful.
	if strings.Contains(err.Error(), `"success": true,`) {
		return nil
	}

	return err
}

func (c *cloudflareProvider) updateSingleRedirect(domainID string, oldrec, newrec *models.RecordConfig) error {
	// rulesetID := cfr.SRRRulesetID
	// rulesetRuleID := cfr.SRRRulesetRuleID
	//printer.Printf("DEBUG: UPDATE-DEL domID=%v sr=%+v\n", domainID, cfr)
	if err := c.deleteSingleRedirects(domainID, *oldrec.CloudflareRedirect); err != nil {
		return err
	}
	//printer.Printf("DEBUG: UPDATE-CREATE domID=%v sr=%+v\n", domainID, newrec.CloudflareRedirect)
	return c.createSingleRedirect(domainID, *newrec.CloudflareRedirect)
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
		code := intZero(value["status_code"])
		raw := fmt.Sprintf("%s,%s,%d,%d", // $FROM,$TO,$PRIO,$CODE
			pr.Targets[0].Constraint.Value,
			value["url"],
			pr.Priority,
			code)
		r.SetTarget(raw)

		cr, err := newCfsrFromUserInput(raw, code, pr.Priority)
		if err != nil {
			return nil, err
		}
		r.CloudflareRedirect = cr

		recs = append(recs, r)
	}
	return recs, nil
}

func (c *cloudflareProvider) deletePageRule(recordID, domainID string) error {
	return c.cfClient.DeletePageRule(context.Background(), domainID, recordID)
}

func (c *cloudflareProvider) updatePageRule(recordID, domainID string, cfr models.CloudflareSingleRedirectConfig) error {
	// maybe someday?
	//c.apiProvider.UpdatePageRule(context.Background(), domainId, recordID, )
	if err := c.deletePageRule(recordID, domainID); err != nil {
		return err
	}
	return c.createPageRule(domainID, cfr)
}

func (c *cloudflareProvider) createPageRule(domainID string, cfr models.CloudflareSingleRedirectConfig) error {
	//printer.Printf("DEBUG: called createPageRule(%s, %+v)\n", domainID, cfr)
	// from to priority code
	// parts := strings.Split(target, ",")
	// priority, _ := strconv.Atoi(parts[2])
	// code, _ := strconv.Atoi(parts[3])
	// printer.Printf("DEBUG: pr.PageRule target = %v\n", target)
	// printer.Printf("DEBUG: pr.PageRule target = %v\n", parts[0])
	// printer.Printf("DEBUG: pr.PageRule url    = %v\n", parts[1])
	// printer.Printf("DEBUG: pr.PageRule code   = %v\n", code)
	priority := cfr.PRPriority
	code := cfr.Code
	matcher := cfr.PRMatcher
	replacement := cfr.PRReplacement
	pr := cloudflare.PageRule{
		Status:   "active",
		Priority: priority,
		Targets: []cloudflare.PageRuleTarget{
			{Target: "url", Constraint: pageRuleConstraint{Operator: "matches", Value: matcher}},
		},
		Actions: []cloudflare.PageRuleAction{
			{ID: "forwarding_url", Value: &pageRuleFwdInfo{
				StatusCode: code,
				URL:        replacement,
			}},
		},
	}
	//printer.Printf("DEBUG: createPageRule pr=%+v\n", pr)
	_, err := c.cfClient.CreatePageRule(context.Background(), domainID, pr)
	return err
}

func (c *cloudflareProvider) getWorkerRoutes(id string, domain string) ([]*models.RecordConfig, error) {
	res, err := c.cfClient.ListWorkerRoutes(context.Background(), cloudflare.ZoneIdentifier(id), cloudflare.ListWorkerRoutesParams{})
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
			pr.ScriptName))
		recs = append(recs, r)
	}
	return recs, nil
}

func (c *cloudflareProvider) deleteWorkerRoute(recordID, domainID string) error {
	_, err := c.cfClient.DeleteWorkerRoute(context.Background(), cloudflare.ZoneIdentifier(domainID), recordID)
	return err
}

func (c *cloudflareProvider) updateWorkerRoute(recordID, domainID string, target string) error {
	// Causing Stack Overflow (!?)
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
	wr := cloudflare.CreateWorkerRouteParams{
		Pattern: parts[0],
		Script:  parts[1],
	}

	_, err := c.cfClient.CreateWorkerRoute(context.Background(), cloudflare.ZoneIdentifier(domainID), wr)
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
