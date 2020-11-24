package cloudflare

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/pkg/transform"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

/*

Cloudflare API DNS provider:

Info required in `creds.json`:
   - apikey
   - apiuser
   - accountid (optional)
   - accountname (optional)

Record level metadata available:
   - cloudflare_proxy ("on", "off", or "full")

Domain level metadata available:
   - cloudflare_proxy_default ("on", "off", or "full")

 Provider level metadata available:
   - ip_conversions
*/

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can("CF automatically flattens CNAME records into A records dynamically"),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot("Cloudflare will not work well in situations where it is not the only DNS server"),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanGetZones:            providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("CLOUDFLAREAPI", newCloudflare, features)
	providers.RegisterCustomRecordType("CF_REDIRECT", "CLOUDFLAREAPI", "")
	providers.RegisterCustomRecordType("CF_TEMP_REDIRECT", "CLOUDFLAREAPI", "")
}

// cloudflareProvider is the handle for API calls.
type cloudflareProvider struct {
	APIKey          string `json:"apikey"`
	APIToken        string `json:"apitoken"`
	APIUser         string `json:"apiuser"`
	AccountID       string `json:"accountid"`
	AccountName     string `json:"accountname"`
	domainIndex     map[string]string
	nameservers     map[string][]string
	ipConversions   []transform.IPConversion
	ignoredLabels   []string
	manageRedirects bool
}

func labelMatches(label string, matches []string) bool {
	printer.Debugf("DEBUG: labelMatches(%#v, %#v)\n", label, matches)
	for _, tst := range matches {
		if label == tst {
			return true
		}
	}
	return false
}

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
func (c *cloudflareProvider) GetZoneRecords(domain string) (models.Records, error) {
	id, err := c.getDomainID(domain)
	if err != nil {
		return nil, err
	}
	records, err := c.getRecordsForDomain(id, domain)
	if err != nil {
		return nil, err
	}
	for _, rec := range records {
		if rec.TTL == 1 {
			rec.TTL = 0
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

// GetDomainCorrections returns a list of corrections to update a domain.
func (c *cloudflareProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	id, err := c.getDomainID(dc.Name)
	if err != nil {
		return nil, err
	}
	records, err := c.getRecordsForDomain(id, dc.Name)
	if err != nil {
		return nil, err
	}

	if err := c.preprocessConfig(dc); err != nil {
		return nil, err
	}
	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		// Delete ignore labels
		if labelMatches(dnsutil.TrimDomainName(rec.Original.(*cfRecord).Name, dc.Name), c.ignoredLabels) {
			printer.Debugf("ignored_label: %s\n", rec.Original.(*cfRecord).Name)
			records = append(records[:i], records[i+1:]...)
		}
	}

	if c.manageRedirects {
		prs, err := c.getPageRules(id, dc.Name)
		if err != nil {
			return nil, err
		}
		records = append(records, prs...)
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
		if labelMatches(rec.GetLabel(), c.ignoredLabels) {
			log.Fatalf("FATAL: dnsconfig contains label that matches ignored_labels: %#v is in %v)\n", rec.GetLabel(), c.ignoredLabels)
		}
	}

	checkNSModifications(dc)

	// Normalize
	models.PostProcessRecords(records)

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
				F:   func() error { return c.deletePageRule(ex.Original.(*pageRule).ID, id) },
			})

		} else {
			corrections = append(corrections, c.deleteRec(ex.Original.(*cfRecord), id))
		}
	}
	for _, d := range create {
		des := d.Desired
		if des.Type == "PAGE_RULE" {
			corrections = append(corrections, &models.Correction{
				Msg: d.String(),
				F:   func() error { return c.createPageRule(id, des.GetTargetField()) },
			})
		} else {
			corrections = append(corrections, c.createRec(des, id)...)
		}
	}

	for _, d := range mod {
		rec := d.Desired
		ex := d.Existing
		if rec.Type == "PAGE_RULE" {
			corrections = append(corrections, &models.Correction{
				Msg: d.String(),
				F:   func() error { return c.updatePageRule(ex.Original.(*pageRule).ID, id, rec.GetTargetField()) },
			})
		} else {
			e := ex.Original.(*cfRecord)
			proxy := e.Proxiable && rec.Metadata[metaProxy] != "off"
			corrections = append(corrections, &models.Correction{
				Msg: d.String(),
				F:   func() error { return c.modifyRecord(id, e.ID, proxy, rec) },
			})
		}
	}

	// Add universalSSL change to corrections when needed
	if changed, newState, err := c.checkUniversalSSL(dc, id); err == nil && changed {
		var newStateString string
		if newState {
			newStateString = "enabled"
		} else {
			newStateString = "disabled"
		}
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("Universal SSL will be %s for this domain.", newStateString),
			F:   func() error { return c.changeUniversalSSL(id, newState) },
		})
	}

	return corrections, nil
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
		if rec.TTL != 1 && rec.TTL < 120 {
			rec.TTL = 120
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
	api.APIUser, api.APIKey, api.APIToken = m["apiuser"], m["apikey"], m["apitoken"]
	// check api keys from creds json file
	if api.APIToken == "" && (api.APIKey == "" || api.APIUser == "") {
		return nil, fmt.Errorf("if cloudflare apitoken is not set, apikey and apiuser must be provided")
	}
	if api.APIToken != "" && (api.APIKey != "" || api.APIUser != "") {
		return nil, fmt.Errorf("if cloudflare apitoken is set, apikey and apiuser should not be provided")
	}

	// Check account data if set
	api.AccountID, api.AccountName = m["accountid"], m["accountname"]
	if (api.AccountID != "" && api.AccountName == "") || (api.AccountID == "" && api.AccountName != "") {
		return nil, fmt.Errorf("either both cloudflare accountid and accountname must be provided or neither")
	}

	err := api.fetchDomainList()
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		parsedMeta := &struct {
			IPConversions   string   `json:"ip_conversions"`
			IgnoredLabels   []string `json:"ignored_labels"`
			ManageRedirects bool     `json:"manage_redirects"`
		}{}
		err := json.Unmarshal([]byte(metadata), parsedMeta)
		if err != nil {
			return nil, err
		}
		api.manageRedirects = parsedMeta.ManageRedirects
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

type cfRecord struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Content    string      `json:"content"`
	Proxiable  bool        `json:"proxiable"`
	Proxied    bool        `json:"proxied"`
	TTL        uint32      `json:"ttl"`
	Locked     bool        `json:"locked"`
	ZoneID     string      `json:"zone_id"`
	ZoneName   string      `json:"zone_name"`
	CreatedOn  time.Time   `json:"created_on"`
	ModifiedOn time.Time   `json:"modified_on"`
	Data       *cfRecData  `json:"data"`
	Priority   json.Number `json:"priority"`
}

func (c *cfRecord) nativeToRecord(domain string) (*models.RecordConfig, error) {
	// normalize cname,mx,ns records with dots to be consistent with our config format.
	if c.Type == "CNAME" || c.Type == "MX" || c.Type == "NS" || c.Type == "SRV" {
		if c.Content != "." {
			c.Content = c.Content + "."
		}
	}

	rc := &models.RecordConfig{
		TTL:      c.TTL,
		Original: c,
	}
	rc.SetLabelFromFQDN(c.Name, domain)

	// workaround for https://github.com/StackExchange/dnscontrol/issues/446
	if c.Type == "SPF" {
		c.Type = "TXT"
	}

	switch rType := c.Type; rType { // #rtype_variations
	case "MX":
		var priority uint16
		if c.Priority == "" {
			priority = 0
		} else {
			p, err := c.Priority.Int64()
			if err != nil {
				return nil, fmt.Errorf("error decoding priority from cloudflare record: %w", err)
			}
			priority = uint16(p)
		}
		if err := rc.SetTargetMX(priority, c.Content); err != nil {
			return nil, fmt.Errorf("unparsable MX record received from cloudflare: %w", err)
		}
	case "SRV":
		data := *c.Data
		if err := rc.SetTargetSRV(data.Priority, data.Weight, data.Port,
			dnsutil.AddOrigin(data.Target.FQDN(), domain)); err != nil {
			return nil, fmt.Errorf("unparsable SRV record received from cloudflare: %w", err)
		}
	default: // "A", "AAAA", "ANAME", "CAA", "CNAME", "NS", "PTR", "TXT"
		if err := rc.PopulateFromString(rType, c.Content, domain); err != nil {
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
		proxied = r.Original.(*cfRecord).Proxied
	} else {
		proxied = r.Metadata[metaProxy] != "off"
	}
	return map[string]string{
		"proxy": fmt.Sprint(proxied),
	}
}

// EnsureDomainExists returns an error of domain does not exist.
func (c *cloudflareProvider) EnsureDomainExists(domain string) error {
	if _, ok := c.domainIndex[domain]; ok {
		return nil
	}
	var id string
	id, err := c.createZone(domain)
	fmt.Printf("Added zone for %s to Cloudflare account: %s\n", domain, id)
	return err
}
