package cloudflare

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/pkg/transform"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/cloudflare/cloudflare-go"
	"github.com/fatih/color"
)

/*

Cloudflare API DNS provider:

Info required in `creds.json`:
   - apikey
   - apiuser
   - accountid (optional)

Record level metadata available:
   - cloudflare_proxy ("on", "off", or "full")

Domain level metadata available:
   - cloudflare_proxy_default ("on", "off", or "full")

 Provider level metadata available:
   - ip_conversions
*/

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can("CF automatically flattens CNAME records into A records dynamically"),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot("Cloudflare will not work well in situations where it is not the only DNS server"),
	providers.DocOfficiallySupported: providers.Can(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newCloudflare,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("CLOUDFLAREAPI", fns, features)
	providers.RegisterCustomRecordType("CF_REDIRECT", "CLOUDFLAREAPI", "")
	providers.RegisterCustomRecordType("CF_TEMP_REDIRECT", "CLOUDFLAREAPI", "")
	providers.RegisterCustomRecordType("CF_WORKER_ROUTE", "CLOUDFLAREAPI", "")
}

// cloudflareProvider is the handle for API calls.
type cloudflareProvider struct {
	domainIndex     map[string]string // Call c.fetchDomainList() to populate before use.
	nameservers     map[string][]string
	ipConversions   []transform.IPConversion
	ignoredLabels   []string
	manageRedirects bool
	manageWorkers   bool
	cfClient        *cloudflare.API
}

// TODO(dlemenkov): remove this function after deleting all commented code referecing it
//func labelMatches(label string, matches []string) bool {
//	printer.Debugf("DEBUG: labelMatches(%#v, %#v)\n", label, matches)
//	for _, tst := range matches {
//		if label == tst {
//			return true
//		}
//	}
//	return false
//}

// GetNameservers returns the nameservers for a domain.
func (c *cloudflareProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	ns, ok := c.nameservers[domain]
	if !ok {
		return nil, fmt.Errorf("nameservers for %s not found in cloudflare account", domain)
	}
	return models.ToNameservers(ns)
}

// ListZones returns a list of the DNS zones.
func (c *cloudflareProvider) ListZones() ([]string, error) {
	if err := c.fetchDomainList(); err != nil {
		return nil, err
	}
	zones := make([]string, 0, len(c.domainIndex))
	for d := range c.domainIndex {
		zones = append(zones, d)
	}
	return zones, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *cloudflareProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {

	domainID, err := c.getDomainID(domain)
	if err != nil {
		return nil, err
	}
	records, err := c.getRecordsForDomain(domainID, domain)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		if rec.TTL == 0 {
			rec.TTL = 1
		}
		// Store the proxy status ("orange cloud") for use by get-zones:
		m := getProxyMetadata(rec)
		if p, ok := m["proxy"]; ok {
			if rec.Metadata == nil {
				rec.Metadata = map[string]string{}
			}
			rec.Metadata["cloudflare_proxy"] = p
		}
	}

	// // FIXME(tlim) Why is this needed???
	// // I don't know. Let's comment it out and see if anything breaks.
	// for i := len(records) - 1; i >= 0; i-- {
	// 	rec := records[i]
	// 	// Delete ignore labels
	// 	if labelMatches(dnsutil.TrimDomainName(rec.Original.(cloudflare.DNSRecord).Name, dc.Name), c.ignoredLabels) {
	// 		printer.Debugf("ignored_label: %s\n", rec.Original.(cloudflare.DNSRecord).Name)
	// 		records = append(records[:i], records[i+1:]...)
	// 	}
	// }

	if c.manageRedirects {
		prs, err := c.getPageRules(domainID, domain)
		if err != nil {
			return nil, err
		}
		records = append(records, prs...)
	}

	if c.manageWorkers {
		wrs, err := c.getWorkerRoutes(domainID, domain)
		if err != nil {
			return nil, err
		}
		records = append(records, wrs...)
	}

	// Normalize
	models.PostProcessRecords(records)

	return records, nil
}

func (c *cloudflareProvider) getDomainID(name string) (string, error) {
	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return "", err
		}
	}
	id, ok := c.domainIndex[name]
	if !ok {
		return "", fmt.Errorf("'%s' not a zone in cloudflare account", name)
	}
	return id, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (c *cloudflareProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, error) {

	if err := c.preprocessConfig(dc); err != nil {
		return nil, err
	}
	//	for i := len(records) - 1; i >= 0; i-- {
	//		rec := records[i]
	//		// Delete ignore labels
	//		if labelMatches(dnsutil.TrimDomainName(rec.Original.(cloudflare.DNSRecord).Name, dc.Name), c.ignoredLabels) {
	//			printer.Debugf("ignored_label: %s\n", rec.Original.(cloudflare.DNSRecord).Name)
	//			records = append(records[:i], records[i+1:]...)
	//		}
	//	}

	checkNSModifications(dc)

	domainID, err := c.getDomainID(dc.Name)
	if err != nil {
		return nil, err
	}

	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" {
			rec.Type = "CNAME"
		}
		// As per CF-API documentation proxied records are always forced to have a TTL of 1.
		// When not forcing this property change here, dnscontrol tries each time to update
		// the TTL of a record which simply cannot be changed anyway.
		if rec.Metadata[metaProxy] != "off" {
			rec.TTL = 1
		}
		//		if labelMatches(rec.GetLabel(), c.ignoredLabels) {
		//			log.Fatalf("FATAL: dnsconfig contains label that matches ignored_labels: %#v is in %v)\n", rec.GetLabel(), c.ignoredLabels)
		//		}
	}

	checkNSModifications(dc)

	var corrections []*models.Correction
	if !diff2.EnableDiff2 {

		differ := diff.New(dc, getProxyMetadata)
		_, create, del, mod, err := differ.IncrementalDiff(records)
		if err != nil {
			return nil, err
		}

		corrections := []*models.Correction{}

		for _, d := range del {
			ex := d.Existing
			if ex.Type == "PAGE_RULE" {
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F:   func() error { return c.deletePageRule(ex.Original.(cloudflare.PageRule).ID, domainID) },
				})
			} else if ex.Type == "WORKER_ROUTE" {
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F:   func() error { return c.deleteWorkerRoute(ex.Original.(cloudflare.WorkerRoute).ID, domainID) },
				})
			} else {
				corr := c.deleteRec(ex.Original.(cloudflare.DNSRecord), domainID)
				// DS records must always have a corresponding NS record.
				// Therefore, we remove DS records before any NS records.
				if d.Existing.Type == "DS" {
					corrections = append([]*models.Correction{corr}, corrections...)
				} else {
					corrections = append(corrections, corr)
				}
			}
		}
		for _, d := range create {
			des := d.Desired
			if des.Type == "PAGE_RULE" {
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F:   func() error { return c.createPageRule(domainID, des.GetTargetField()) },
				})
			} else if des.Type == "WORKER_ROUTE" {
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F:   func() error { return c.createWorkerRoute(domainID, des.GetTargetField()) },
				})
			} else {
				corr := c.createRec(des, domainID)
				// DS records must always have a corresponding NS record.
				// Therefore, we create NS records before any DS records.
				if d.Desired.Type == "NS" {
					corrections = append(corr, corrections...)
				} else {
					corrections = append(corrections, corr...)
				}
			}
		}

		for _, d := range mod {
			rec := d.Desired
			ex := d.Existing
			if rec.Type == "PAGE_RULE" {
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F: func() error {
						return c.updatePageRule(ex.Original.(cloudflare.PageRule).ID, domainID, rec.GetTargetField())
					},
				})
			} else if rec.Type == "WORKER_ROUTE" {
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F: func() error {
						return c.updateWorkerRoute(ex.Original.(cloudflare.WorkerRoute).ID, domainID, rec.GetTargetField())
					},
				})
			} else {
				e := ex.Original.(cloudflare.DNSRecord)
				proxy := e.Proxiable && rec.Metadata[metaProxy] != "off"
				corrections = append(corrections, &models.Correction{
					Msg: d.String(),
					F:   func() error { return c.modifyRecord(domainID, e.ID, proxy, rec) },
				})
			}
		}

		// Add universalSSL change to corrections when needed
		if changed, newState, err := c.checkUniversalSSL(dc, domainID); err == nil && changed {
			var newStateString string
			if newState {
				newStateString = "enabled"
			} else {
				newStateString = "disabled"
			}
			corrections = append(corrections, &models.Correction{
				Msg: fmt.Sprintf("Universal SSL will be %s for this domain.", newStateString),
				F:   func() error { return c.changeUniversalSSL(domainID, newState) },
			})
		}

		return corrections, nil
	}

	// Cloudflare is a "ByRecord" API.
	instructions, err := diff2.ByRecord(records, dc, genComparable)
	if err != nil {
		return nil, err
	}

	for _, inst := range instructions {

		addToFront := false
		var corrs []*models.Correction

		domainID := domainID
		msg := inst.Msgs[0]

		switch inst.Type {
		case diff2.CREATE:
			createRec := inst.New[0]
			corrs = c.mkCreateCorrection(createRec, domainID, msg)
			// DS records must always have a corresponding NS record.
			// Therefore, we create NS records before any DS records.
			addToFront = (createRec.Type == "NS")
		case diff2.CHANGE:
			newrec := inst.New[0]
			oldrec := inst.Old[0]
			corrs = c.mkChangeCorrection(oldrec, newrec, domainID, msg)
		case diff2.DELETE:
			deleteRec := inst.Old[0]
			deleteRecType := deleteRec.Type
			deleteRecOrig := deleteRec.Original
			corrs = c.mkDeleteCorrection(deleteRecType, deleteRecOrig, domainID, msg)
			// DS records must always have a corresponding NS record.
			// Therefore, we remove DS records before any NS records.
			addToFront = (deleteRecType == "DS")
		}

		if addToFront {
			corrections = append(corrs, corrections...)
		} else {
			corrections = append(corrections, corrs...)
		}
	}

	// Add universalSSL change when needed
	if changed, newState, err := c.checkUniversalSSL(dc, domainID); err == nil && changed {
		var newStateString string
		if newState {
			newStateString = "enabled"
		} else {
			newStateString = "disabled"
		}
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("Universal SSL will be %s for this domain.", newStateString),
			F:   func() error { return c.changeUniversalSSL(domainID, newState) },
		})
	}

	return corrections, nil
}

func genComparable(rec *models.RecordConfig) string {
	if rec.Type == "A" || rec.Type == "AAAA" || rec.Type == "CNAME" {
		proxy := rec.Metadata[metaProxy]
		if proxy != "" {
			if proxy == "on" {
				proxy = "true"
			}
			if proxy == "off" {
				proxy = "false"
			}
			return "proxy=" + proxy
		}
	}
	return ""
}

func (c *cloudflareProvider) mkCreateCorrection(newrec *models.RecordConfig, domainID, msg string) []*models.Correction {
	switch newrec.Type {
	case "PAGE_RULE":
		return []*models.Correction{{
			Msg: msg,
			F:   func() error { return c.createPageRule(domainID, newrec.GetTargetField()) },
		}}
	case "WORKER_ROUTE":
		return []*models.Correction{{
			Msg: msg,
			F:   func() error { return c.createWorkerRoute(domainID, newrec.GetTargetField()) },
		}}
	default:
		return c.createRecDiff2(newrec, domainID, msg)
	}
}

func (c *cloudflareProvider) mkChangeCorrection(oldrec, newrec *models.RecordConfig, domainID string, msg string) []*models.Correction {

	var idTxt string
	switch oldrec.Type {
	case "PAGE_RULE":
		idTxt = oldrec.Original.(cloudflare.PageRule).ID
	case "WORKER_ROUTE":
		idTxt = oldrec.Original.(cloudflare.WorkerRoute).ID
	default:
		idTxt = oldrec.Original.(cloudflare.DNSRecord).ID
	}
	msg = msg + color.YellowString(" id=%v", idTxt)

	switch newrec.Type {
	case "PAGE_RULE":
		return []*models.Correction{{
			Msg: msg,
			F: func() error {
				return c.updatePageRule(idTxt, domainID, newrec.GetTargetField())
			},
		}}
	case "WORKER_ROUTE":
		return []*models.Correction{{
			Msg: msg,
			F: func() error {
				return c.updateWorkerRoute(idTxt, domainID, newrec.GetTargetField())
			},
		}}
	default:
		e := oldrec.Original.(cloudflare.DNSRecord)
		proxy := e.Proxiable && newrec.Metadata[metaProxy] != "off"
		return []*models.Correction{{
			Msg: msg,
			F:   func() error { return c.modifyRecord(domainID, e.ID, proxy, newrec) },
		}}
	}
}

func (c *cloudflareProvider) mkDeleteCorrection(recType string, origRec any, domainID string, msg string) []*models.Correction {

	var idTxt string
	switch recType {
	case "PAGE_RULE":
		idTxt = origRec.(cloudflare.PageRule).ID
	case "WORKER_ROUTE":
		idTxt = origRec.(cloudflare.WorkerRoute).ID
	default:
		idTxt = origRec.(cloudflare.DNSRecord).ID
	}
	msg = msg + color.RedString(" id=%v", idTxt)

	correction := &models.Correction{
		Msg: msg,
		F: func() error {
			switch recType {
			case "PAGE_RULE":
				return c.deletePageRule(origRec.(cloudflare.PageRule).ID, domainID)
			case "WORKER_ROUTE":
				return c.deleteWorkerRoute(origRec.(cloudflare.WorkerRoute).ID, domainID)
			default:
				return c.deleteDNSRecord(origRec.(cloudflare.DNSRecord), domainID)
			}
		},
	}
	return []*models.Correction{correction}
}

func checkNSModifications(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.GetLabelFQDN() == dc.Name {
			if !strings.HasSuffix(rec.GetTargetField(), ".ns.cloudflare.com.") {
				printer.Warnf("cloudflare does not support modifying NS records on base domain. %s will not be added.\n", rec.GetTargetField())
			}
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

func (c *cloudflareProvider) checkUniversalSSL(dc *models.DomainConfig, id string) (changed bool, newState bool, err error) {
	expectedStr := dc.Metadata[metaUniversalSSL]
	if expectedStr == "" {
		return false, false, fmt.Errorf("metadata not set")
	}

	if actual, err := c.getUniversalSSL(id); err == nil {
		// convert str to bool
		var expected bool
		if expectedStr == "off" {
			expected = false
		} else {
			expected = true
		}
		// did something change?
		if actual != expected {
			return true, expected, nil
		}
		return false, expected, nil
	}
	return false, false, fmt.Errorf("error receiving universal ssl state")
}

const (
	metaProxy         = "cloudflare_proxy"
	metaProxyDefault  = metaProxy + "_default"
	metaOriginalIP    = "original_ip" // TODO(tlim): Unclear what this means.
	metaUniversalSSL  = "cloudflare_universalssl"
	metaIPConversions = "ip_conversions" // TODO(tlim): Rename to obscure_rules.
)

func checkProxyVal(v string) (string, error) {
	v = strings.ToLower(v)
	if v != "on" && v != "off" && v != "full" {
		return "", fmt.Errorf("bad metadata value for cloudflare_proxy: '%s'. Use on/off/full", v)
	}
	return v, nil
}

func (c *cloudflareProvider) preprocessConfig(dc *models.DomainConfig) error {

	// Determine the default proxy setting.
	var defProxy string
	var err error
	if defProxy = dc.Metadata[metaProxyDefault]; defProxy == "" {
		defProxy = "off"
	} else {
		defProxy, err = checkProxyVal(defProxy)
		if err != nil {
			return err
		}
	}

	// Check UniversalSSL setting
	if u := dc.Metadata[metaUniversalSSL]; u != "" {
		u = strings.ToLower(u)
		if u != "on" && u != "off" {
			return fmt.Errorf("bad metadata value for %s: '%s'. Use on/off", metaUniversalSSL, u)
		}
	}

	// Normalize the proxy setting for each record.
	// A and CNAMEs: Validate. If null, set to default.
	// else: Make sure it wasn't set.  Set to default.
	// iterate backwards so first defined page rules have highest priority
	currentPrPrio := 1
	for i := len(dc.Records) - 1; i >= 0; i-- {
		rec := dc.Records[i]
		if rec.Metadata == nil {
			rec.Metadata = map[string]string{}
		}
		// cloudflare uses "1" to mean "auto-ttl"
		// if we get here and ttl is not specified (or is the dnscontrol default of 300),
		// use automatic mode instead.
		if rec.TTL == 0 || rec.TTL == 300 {
			rec.TTL = 1
		}
		if rec.TTL != 1 && rec.TTL < 60 {
			rec.TTL = 60
		}

		if rec.Type != "A" && rec.Type != "CNAME" && rec.Type != "AAAA" && rec.Type != "ALIAS" {
			if rec.Metadata[metaProxy] != "" {
				return fmt.Errorf("cloudflare_proxy set on %v record: %#v cloudflare_proxy=%#v", rec.Type, rec.GetLabel(), rec.Metadata[metaProxy])
			}
			// Force it to off.
			rec.Metadata[metaProxy] = "off"
		} else {
			if val := rec.Metadata[metaProxy]; val == "" {
				rec.Metadata[metaProxy] = defProxy
			} else {
				val, err := checkProxyVal(val)
				if err != nil {
					return err
				}
				rec.Metadata[metaProxy] = val
			}
		}

		// CF_REDIRECT record types. Encode target as $FROM,$TO,$PRIO,$CODE
		if rec.Type == "CF_REDIRECT" || rec.Type == "CF_TEMP_REDIRECT" {
			if !c.manageRedirects {
				return fmt.Errorf("you must add 'manage_redirects: true' metadata to cloudflare provider to use CF_REDIRECT records")
			}
			parts := strings.Split(rec.GetTargetField(), ",")
			if len(parts) != 2 {
				return fmt.Errorf("invalid data specified for cloudflare redirect record")
			}
			code := 301
			if rec.Type == "CF_TEMP_REDIRECT" {
				code = 302
			}
			rec.SetTarget(fmt.Sprintf("%s,%d,%d", rec.GetTargetField(), currentPrPrio, code))
			currentPrPrio++
			rec.TTL = 1
			rec.Type = "PAGE_RULE"
		}

		// CF_WORKER_ROUTE record types. Encode target as $PATTERN,$SCRIPT
		if rec.Type == "CF_WORKER_ROUTE" {
			parts := strings.Split(rec.GetTargetField(), ",")
			if len(parts) != 2 {
				return fmt.Errorf("invalid data specified for cloudflare worker record")
			}
			rec.TTL = 1
			rec.Type = "WORKER_ROUTE"
		}
	}

	// look for ip conversions and transform records
	for _, rec := range dc.Records {
		if rec.Type != "A" {
			continue
		}
		// only transform "full"
		if rec.Metadata[metaProxy] != "full" {
			continue
		}
		ip := net.ParseIP(rec.GetTargetField())
		if ip == nil {
			return fmt.Errorf("%s is not a valid ip address", rec.GetTargetField())
		}
		newIP, err := transform.IP(ip, c.ipConversions)
		if err != nil {
			return err
		}
		rec.Metadata[metaOriginalIP] = rec.GetTargetField()
		rec.SetTarget(newIP.String())
	}

	return nil
}

func newCloudflare(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &cloudflareProvider{}
	// check api keys from creds json file
	if m["apitoken"] == "" && (m["apikey"] == "" || m["apiuser"] == "") {
		return nil, fmt.Errorf("if cloudflare apitoken is not set, apikey and apiuser must be provided")
	}
	if m["apitoken"] != "" && (m["apikey"] != "" || m["apiuser"] != "") {
		return nil, fmt.Errorf("if cloudflare apitoken is set, apikey and apiuser should not be provided")
	}

	optRP := cloudflare.UsingRetryPolicy(20, 1, 120)
	// UsingRetryPolicy is documented here:
	// https://pkg.go.dev/github.com/cloudflare/cloudflare-go#UsingRetryPolicy
	// The defaults are UsingRetryPolicy(3, 1, 30)

	var err error
	if m["apitoken"] != "" {
		api.cfClient, err = cloudflare.NewWithAPIToken(m["apitoken"], optRP)
	} else {
		api.cfClient, err = cloudflare.New(m["apikey"], m["apiuser"], optRP)
	}

	if err != nil {
		return nil, fmt.Errorf("cloudflare credentials: %w", err)
	}

	// Check account data if set
	if m["accountid"] != "" {
		api.cfClient.AccountID = m["accountid"]
	}

	if len(metadata) > 0 {
		parsedMeta := &struct {
			IPConversions   string   `json:"ip_conversions"`
			IgnoredLabels   []string `json:"ignored_labels"`
			ManageRedirects bool     `json:"manage_redirects"`
			ManageWorkers   bool     `json:"manage_workers"`
		}{}
		err := json.Unmarshal([]byte(metadata), parsedMeta)
		if err != nil {
			return nil, err
		}
		api.manageRedirects = parsedMeta.ManageRedirects
		api.manageWorkers = parsedMeta.ManageWorkers
		// ignored_labels:
		api.ignoredLabels = append(api.ignoredLabels, parsedMeta.IgnoredLabels...)
		if len(api.ignoredLabels) > 0 {
			printer.Warnf("Cloudflare 'ignored_labels' configuration is deprecated and might be removed. Please use the IGNORE domain directive to achieve the same effect.\n")
		}
		// parse provider level metadata
		if len(parsedMeta.IPConversions) > 0 {
			api.ipConversions, err = transform.DecodeTransformTable(parsedMeta.IPConversions)
			if err != nil {
				return nil, err
			}
		}
	}
	return api, nil
}

// Used on the "existing" records.
type cfRecData struct {
	Name         string   `json:"name"`
	Target       cfTarget `json:"target"`
	Service      string   `json:"service"`       // SRV
	Proto        string   `json:"proto"`         // SRV
	Priority     uint16   `json:"priority"`      // SRV
	Weight       uint16   `json:"weight"`        // SRV
	Port         uint16   `json:"port"`          // SRV
	Tag          string   `json:"tag"`           // CAA
	Flags        uint8    `json:"flags"`         // CAA
	Value        string   `json:"value"`         // CAA
	Usage        uint8    `json:"usage"`         // TLSA
	Selector     uint8    `json:"selector"`      // TLSA
	MatchingType uint8    `json:"matching_type"` // TLSA
	Certificate  string   `json:"certificate"`   // TLSA
	Algorithm    uint8    `json:"algorithm"`     // SSHFP/DS
	HashType     uint8    `json:"type"`          // SSHFP
	Fingerprint  string   `json:"fingerprint"`   // SSHFP
	KeyTag       uint16   `json:"key_tag"`       // DS
	DigestType   uint8    `json:"digest_type"`   // DS
	Digest       string   `json:"digest"`        // DS
}

// cfTarget is a SRV target. A null target is represented by an empty string, but
// a dot is so acceptable.
type cfTarget string

// UnmarshalJSON decodes a SRV target from the Cloudflare API. A null target is
// represented by a false boolean or a dot. Domain names are FQDNs without a
// trailing period (as of 2019-11-05).
func (c *cfTarget) UnmarshalJSON(data []byte) error {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	switch v := obj.(type) {
	case string:
		*c = cfTarget(v)
	case bool:
		if v {
			panic("unknown value for cfTarget bool: true")
		}
		*c = "" // the "." is already added by nativeToRecord
	}
	return nil
}

// MarshalJSON encodes cfTarget for the Cloudflare API. Null targets are
// represented by a single period.
func (c cfTarget) MarshalJSON() ([]byte, error) {
	var obj string
	switch c {
	case "", ".":
		obj = "."
	default:
		obj = string(c)
	}
	return json.Marshal(obj)
}

// DNSControlString returns cfTarget normalized to be a FQDN. Null targets are
// represented by a single period.
func (c cfTarget) FQDN() string {
	return strings.TrimRight(string(c), ".") + "."
}

// uint16Zero converts value to uint16 or returns 0.
func uint16Zero(value interface{}) uint16 {
	switch v := value.(type) {
	case float64:
		return uint16(v)
	case uint16:
		return v
	case nil:
	}
	return 0
}

// intZero converts value to int or returns 0.
func intZero(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case nil:
	}
	return 0
}

// stringDefault returns the value as a string or returns the default value if nil.
func stringDefault(value interface{}, def string) string {
	switch v := value.(type) {
	case string:
		return v
	case nil:
	}
	return def
}

func (c *cloudflareProvider) nativeToRecord(domain string, cr cloudflare.DNSRecord) (*models.RecordConfig, error) {

	// normalize cname,mx,ns records with dots to be consistent with our config format.
	if cr.Type == "CNAME" || cr.Type == "MX" || cr.Type == "NS" || cr.Type == "PTR" {
		if cr.Content != "." {
			cr.Content = cr.Content + "."
		}
	}

	rc := &models.RecordConfig{
		TTL:      uint32(cr.TTL),
		Original: cr,
		Metadata: map[string]string{},
	}
	rc.SetLabelFromFQDN(cr.Name, domain)

	// workaround for https://github.com/StackExchange/dnscontrol/issues/446
	if cr.Type == "SPF" {
		cr.Type = "TXT"
	}

	if cr.Type == "A" || cr.Type == "AAAA" || cr.Type == "CNAME" {
		if cr.Proxied != nil {
			if *(cr.Proxied) {
				rc.Metadata[metaProxy] = "on"
			} else {
				rc.Metadata[metaProxy] = "off"
			}
		}
	}

	switch rType := cr.Type; rType { // #rtype_variations
	case "MX":
		if err := rc.SetTargetMX(*cr.Priority, cr.Content); err != nil {
			return nil, fmt.Errorf("unparsable MX record received from cloudflare: %w", err)
		}
	case "SRV":
		data := cr.Data.(map[string]interface{})

		target := stringDefault(data["target"], "MISSING.TARGET")
		if target != "." {
			target += "."
		}
		if err := rc.SetTargetSRV(uint16Zero(data["priority"]), uint16Zero(data["weight"]), uint16Zero(data["port"]),
			target); err != nil {
			return nil, fmt.Errorf("unparsable SRV record received from cloudflare: %w", err)
		}
	case "TXT":
		err := rc.SetTargetTXT(cr.Content)
		return rc, err
	default:
		if err := rc.PopulateFromString(rType, cr.Content, domain); err != nil {
			return nil, fmt.Errorf("unparsable record received from cloudflare: %w", err)
		}
	}

	return rc, nil
}

func getProxyMetadata(r *models.RecordConfig) map[string]string {
	if r.Type != "A" && r.Type != "AAAA" && r.Type != "CNAME" {
		return nil
	}
	var proxied bool
	if r.Original != nil {
		proxied = *r.Original.(cloudflare.DNSRecord).Proxied
	} else {
		proxied = r.Metadata[metaProxy] != "off"
	}
	return map[string]string{
		"proxy": fmt.Sprint(proxied),
	}
}

// EnsureZoneExists creates a zone if it does not exist
func (c *cloudflareProvider) EnsureZoneExists(domain string) error {
	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return err
		}
	}
	if _, ok := c.domainIndex[domain]; ok {
		return nil
	}
	var id string
	id, err := c.createZone(domain)
	printer.Printf("Added zone for %s to Cloudflare account: %s\n", domain, id)
	return err
}

// PrepareCloudflareTestWorkers creates Cloudflare Workers required for CF_WORKER_ROUTE tests.
func PrepareCloudflareTestWorkers(prv providers.DNSServiceProvider) error {
	cf, ok := prv.(*cloudflareProvider)
	if ok {
		err := cf.createTestWorker("dnscontrol_integrationtest_cnn")
		if err != nil {
			return err
		}

		err = cf.createTestWorker("dnscontrol_integrationtest_msnbc")
		if err != nil {
			return err
		}
	}
	return nil
}
