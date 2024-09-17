package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/idna"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/pkg/transform"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/cfsingleredirect"
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
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Can(),
	providers.CanUseAlias:            providers.Can("CF automatically flattens CNAME records into A records dynamically"),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDNSKEY:           providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseHTTPS:            providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseSVCB:             providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot("Cloudflare will not work well in situations where it is not the only DNS server"),
	providers.DocOfficiallySupported: providers.Can(),
}

func init() {
	const providerName = "CLOUDFLAREAPI"
	const providerMaintainer = "@tresni"
	fns := providers.DspFuncs{
		Initializer:   newCloudflare,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterCustomRecordType("CF_REDIRECT", providerName, "")
	providers.RegisterCustomRecordType("CF_TEMP_REDIRECT", providerName, "")
	providers.RegisterCustomRecordType("CF_WORKER_ROUTE", providerName, "")
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

// cloudflareProvider is the handle for API calls.
type cloudflareProvider struct {
	ipConversions   []transform.IPConversion
	ignoredLabels   []string
	manageRedirects bool // Old "Page Rule"-style redirects.
	manageWorkers   bool
	accountID       string
	cfClient        *cloudflare.API
	//
	manageSingleRedirects bool // New "Single Redirects"-style redirects.
	//
	// Used by
	tcLogFilename string   // Transcode Log file name
	tcLogFh       *os.File // Transcode Log file handle
	tcZone        string   // Transcode Current zone

	sync.Mutex                      // Protects all access to the following fields:
	domainIndex map[string]string   // Cache of zone name to zone ID.
	nameservers map[string][]string // Cache of zone name to list of nameservers.
}

// GetNameservers returns the nameservers for a domain.
func (c *cloudflareProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {

	c.Lock()
	defer c.Unlock()
	if err := c.cacheDomainList(); err != nil {
		return nil, err
	}

	ns, ok := c.nameservers[domain]
	if !ok {
		return nil, fmt.Errorf("nameservers for %s not found in cloudflare cache(%q)", domain, c.accountID)
	}
	return models.ToNameservers(ns)
}

// ListZones returns a list of the DNS zones.
func (c *cloudflareProvider) ListZones() ([]string, error) {

	c.Lock()
	defer c.Unlock()
	if err := c.cacheDomainList(); err != nil {
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

	if c.manageRedirects { // if old
		prs, err := c.getPageRules(domainID, domain)
		if err != nil {
			return nil, err
		}
		records = append(records, prs...)
	}

	if c.manageSingleRedirects { // if new xor old
		// Download the list of Single Redirects.
		// For each one, generate a SINGLEREDIRECT record
		prs, err := c.getSingleRedirects(domainID, domain)
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

	c.Lock()
	defer c.Unlock()
	if err := c.cacheDomainList(); err != nil {
		return "", err
	}

	id, ok := c.domainIndex[name]
	if !ok {
		return "", fmt.Errorf("'%s' not a zone in cloudflare account", name)
	}
	return id, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (c *cloudflareProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {

	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" {
			rec.Type = "CNAME"
		}
	}

	if err := c.preprocessConfig(dc); err != nil {
		return nil, 0, err
	}

	checkNSModifications(dc)

	domainID, err := c.getDomainID(dc.Name)
	if err != nil {
		return nil, 0, err
	}

	for _, rec := range dc.Records {
		// As per CF-API documentation proxied records are always forced to have a TTL of 1.
		// When not forcing this property change here, dnscontrol tries each time to update
		// the TTL of a record which simply cannot be changed anyway.
		if rec.Metadata[metaProxy] != "off" {
			rec.TTL = 1
		}
	}

	checkNSModifications(dc)

	var corrections []*models.Correction

	// Cloudflare is a "ByRecord" API.
	instructions, actualChangeCount, err := diff2.ByRecord(records, dc, genComparable)
	if err != nil {
		return nil, 0, err
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
			corrs = c.mkDeleteCorrection(deleteRecType, deleteRec, domainID, msg)
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

	return corrections, actualChangeCount, nil
}

func genComparable(rec *models.RecordConfig) string {
	if rec.Type == "A" || rec.Type == "AAAA" || rec.Type == "CNAME" {
		proxy := rec.Metadata[metaProxy]
		if proxy != "" {
			if proxy == "on" || proxy == "full" {
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
			F:   func() error { return c.createPageRule(domainID, *newrec.CloudflareRedirect) },
		}}
	case "WORKER_ROUTE":
		return []*models.Correction{{
			Msg: msg,
			F:   func() error { return c.createWorkerRoute(domainID, newrec.GetTargetField()) },
		}}
	case cfsingleredirect.SINGLEREDIRECT:
		return []*models.Correction{{
			Msg: msg,
			F: func() error {
				return c.createSingleRedirect(domainID, *newrec.CloudflareRedirect)
			},
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
	case cfsingleredirect.SINGLEREDIRECT:
		idTxt = oldrec.CloudflareRedirect.SRRRulesetID
	default:
		idTxt = oldrec.Original.(cloudflare.DNSRecord).ID
	}
	msg = msg + color.YellowString(" id=%v", idTxt)

	switch newrec.Type {
	case "PAGE_RULE":
		return []*models.Correction{{
			Msg: msg,
			F: func() error {
				return c.updatePageRule(idTxt, domainID, *newrec.CloudflareRedirect)
			},
		}}
	case cfsingleredirect.SINGLEREDIRECT:
		return []*models.Correction{{
			Msg: msg,
			F: func() error {
				return c.updateSingleRedirect(domainID, oldrec, newrec)
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
		//fmt.Fprintf(os.Stderr, "DEBUG: proxy := %v && %v != off is... %v\n", e.Proxiable, newrec.Metadata[metaProxy], proxy)
		return []*models.Correction{{
			Msg: msg,
			F:   func() error { return c.modifyRecord(domainID, e.ID, proxy, newrec) },
		}}
	}
}

func (c *cloudflareProvider) mkDeleteCorrection(recType string, origRec *models.RecordConfig, domainID string, msg string) []*models.Correction {

	var idTxt string
	switch recType {
	case "PAGE_RULE":
		idTxt = origRec.Original.(cloudflare.PageRule).ID
	case "WORKER_ROUTE":
		idTxt = origRec.Original.(cloudflare.WorkerRoute).ID
	case cfsingleredirect.SINGLEREDIRECT:
		idTxt = origRec.Original.(cloudflare.RulesetRule).ID
	default:
		idTxt = origRec.Original.(cloudflare.DNSRecord).ID
	}
	msg = msg + color.RedString(" id=%v", idTxt)

	correction := &models.Correction{
		Msg: msg,
		F: func() error {
			switch recType {
			case "PAGE_RULE":
				return c.deletePageRule(origRec.Original.(cloudflare.PageRule).ID, domainID)
			case "WORKER_ROUTE":
				return c.deleteWorkerRoute(origRec.Original.(cloudflare.WorkerRoute).ID, domainID)
			case cfsingleredirect.SINGLEREDIRECT:
				return c.deleteSingleRedirects(domainID, *origRec.CloudflareRedirect)
			default:
				return c.deleteDNSRecord(origRec.Original.(cloudflare.DNSRecord), domainID)
			}
		},
	}
	return []*models.Correction{correction}
}

func checkNSModifications(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))

	punyRoot, err := idna.ToASCII(dc.Name)
	if err != nil {
		punyRoot = dc.Name
	}

	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.GetLabelFQDN() == punyRoot {
			if strings.HasSuffix(rec.GetTargetField(), ".ns.cloudflare.com.") {
				continue
			}
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
	prPriority := 0
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

		// CF_REDIRECT record types:
		if rec.Type == "CF_REDIRECT" || rec.Type == "CF_TEMP_REDIRECT" {
			if !c.manageRedirects && !c.manageSingleRedirects {
				return fmt.Errorf("you must add 'manage_single_redirects: true' metadata to cloudflare provider to use CF_REDIRECT/CF_TEMP_REDIRECT records")
			}
			code := uint16(301)
			if rec.Type == "CF_TEMP_REDIRECT" {
				code = 302
			}

			part := strings.SplitN(rec.GetTargetField(), ",", 2)
			prWhen, prThen := part[0], part[1]
			prPriority++

			// Convert this record to a PAGE_RULE.
			cfsingleredirect.MakePageRule(rec, prPriority, code, prWhen, prThen)
			rec.SetLabel("@", dc.Name)

			if c.manageRedirects && !c.manageSingleRedirects {
				// Old-Style only.  No additional work needed.

			} else if !c.manageRedirects && c.manageSingleRedirects {
				// New-Style only.  Convert PAGE_RULE to SINGLEREDIRECT.
				cfsingleredirect.TranscodePRtoSR(rec)
				if err := c.LogTranscode(dc.Name, rec.CloudflareRedirect); err != nil {
					return err
				}

			} else {
				// Both old-style and new-style enabled!
				// Retain the PAGE_RULE and append an additional SINGLEREDIRECT.

				// make a copy:
				newRec, err := rec.Copy()
				if err != nil {
					return err
				}
				// The copy becomes the CF SingleRedirect
				cfsingleredirect.TranscodePRtoSR(rec)
				if err := c.LogTranscode(dc.Name, rec.CloudflareRedirect); err != nil {
					return err
				}
				// Append the copy to the end of the list.
				dc.Records = append(dc.Records, newRec)

				// The original PAGE_RULE remains untouched.
			}

		} else if rec.Type == cfsingleredirect.SINGLEREDIRECT {
			// SINGLEREDIRECT record types. Verify they are enabled.
			if !c.manageSingleRedirects {
				return fmt.Errorf("you must add 'manage_single_redirects: true' metadata to cloudflare provider to use CF_SINGLE__REDIRECT records")
			}

		} else if rec.Type == "CF_WORKER_ROUTE" {
			// CF_WORKER_ROUTE record types. Encode target as $PATTERN,$SCRIPT
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
		// Only transform A records
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

func (c *cloudflareProvider) LogTranscode(zone string, redirect *models.CloudflareSingleRedirectConfig) error {
	// No filename? Don't log anything.
	filename := c.tcLogFilename
	if filename == "" {
		return nil
	}

	// File not opened already? Open it.
	if c.tcLogFh == nil {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		c.tcLogFh = f
	}
	fh := c.tcLogFh

	// Output "D(zone)"  if needed.
	var text string
	if c.tcZone != zone {
		text = fmt.Sprintf("D(%q, ...\n", zone)
	}
	c.tcZone = zone

	// Generate the new command and output.
	text = text + fmt.Sprintf("    CF_SINGLE_REDIRECT(%q,\n                       %03d,\n                       '%s',\n                       '%s'\n    ),\n",
		redirect.SRName, redirect.Code,
		redirect.SRWhen, redirect.SRThen)
	_, err := fh.WriteString(text)
	return err
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
		api.accountID = m["accountid"]
	}

	debug, err := strconv.ParseBool(os.Getenv("CLOUDFLAREAPI_DEBUG"))
	if err == nil {
		api.cfClient.Debug = debug
	}

	if len(metadata) > 0 {
		parsedMeta := &struct {
			IPConversions   string   `json:"ip_conversions"`
			IgnoredLabels   []string `json:"ignored_labels"`
			ManageRedirects bool     `json:"manage_redirects"` // Old-style PAGE_RULE-based redirects
			ManageWorkers   bool     `json:"manage_workers"`
			//
			ManageSingleRedirects bool   `json:"manage_single_redirects"` // New-style Dynamic "Single Redirects"
			TranscodeLogFilename  string `json:"transcode_log"`           // Log the PAGE_RULE conversions.
		}{}
		err := json.Unmarshal([]byte(metadata), parsedMeta)
		if err != nil {
			return nil, err
		}
		api.manageSingleRedirects = parsedMeta.ManageSingleRedirects
		api.manageRedirects = parsedMeta.ManageRedirects
		api.tcLogFilename = parsedMeta.TranscodeLogFilename
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
	Flags        uint16   `json:"flags"`         // CAA/DNSKEY
	Value        string   `json:"value"`         // CAA
	Usage        uint8    `json:"usage"`         // TLSA
	Selector     uint8    `json:"selector"`      // TLSA
	MatchingType uint8    `json:"matching_type"` // TLSA
	Certificate  string   `json:"certificate"`   // TLSA
	Algorithm    uint8    `json:"algorithm"`     // SSHFP/DNSKEY/DS
	HashType     uint8    `json:"type"`          // SSHFP
	Fingerprint  string   `json:"fingerprint"`   // SSHFP
	Protocol     uint8    `json:"protocol"`      // DNSKEY
	PublicKey    string   `json:"public_key"`    // DNSKEY
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

type cfNaptrRecData struct {
	Flags       string `json:"flags"`
	Order       uint16 `json:"order"`
	Preference  uint16 `json:"preference"`
	Regex       string `json:"regex"`
	Replacement string `json:"replacement"`
	Service     string `json:"service"`
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

// intZero converts value to uint16 or returns 0.
func intZero(value interface{}) uint16 {
	switch v := value.(type) {
	case float64:
		return uint16(v)
	case int:
		return uint16(v)
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
	if cr.Type == "ALIAS" || cr.Type == "CNAME" || cr.Type == "MX" || cr.Type == "NS" || cr.Type == "PTR" {
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

	c.Lock()
	defer c.Unlock()
	if err := c.cacheDomainList(); err != nil {
		return err
	}

	if _, ok := c.domainIndex[domain]; ok {
		return nil
	}
	var id string
	id, err := c.createZone(domain)
	printer.Printf("Added zone for %s to Cloudflare account: %s\n", domain, id)
	clear(c.domainIndex) // clear the cache so that the next caller has to refresh it, thus loading the new ID.
	return err
}

// PrepareCloudflareTestWorkers creates Cloudflare Workers required for CF_WORKER_ROUTE integration tests.
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

func (c *cloudflareProvider) createTestWorker(workerName string) error {
	wp := cloudflare.CreateWorkerParams{
		ScriptName: workerName,
		Script: `
			addEventListener("fetch", (event) => {
				event.respondWith(
					new Response("Ok.", { status: 200 })
				);
			});`,
	}

	_, err := c.cfClient.UploadWorker(context.Background(), cloudflare.AccountIdentifier(c.accountID), wp)
	return err
}
