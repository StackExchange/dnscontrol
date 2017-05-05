package cloudflare

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/StackExchange/dnscontrol/transform"
	"github.com/miekg/dns/dnsutil"
)

/*

Cloudflare API DNS provider:

Info required in `creds.json`:
   - apikey
   - apiuser

Record level metadata available:
   - cloudflare_proxy ("on", "off", or "full")

Domain level metadata available:
   - cloudflare_proxy_default ("on", "off", or "full")

 Provider level metadata available:
   - ip_conversions
*/

type CloudflareApi struct {
	ApiKey        string `json:"apikey"`
	ApiUser       string `json:"apiuser"`
	domainIndex   map[string]string
	nameservers   map[string][]string
	ipConversions []transform.IpConversion
	ignoredLabels []string
}

func labelMatches(label string, matches []string) bool {
	//log.Printf("DEBUG: labelMatches(%#v, %#v)\n", label, matches)
	for _, tst := range matches {
		if label == tst {
			return true
		}
	}
	return false
}
func (c *CloudflareApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	ns, ok := c.nameservers[domain]
	if !ok {
		return nil, fmt.Errorf("Nameservers for %s not found in cloudflare account", domain)
	}
	return models.StringsToNameservers(ns), nil
}

func (c *CloudflareApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	id, ok := c.domainIndex[dc.Name]
	if !ok {
		return nil, fmt.Errorf("%s not listed in zones for cloudflare account", dc.Name)
	}
	if err := c.preprocessConfig(dc); err != nil {
		return nil, err
	}
	records, err := c.getRecordsForDomain(id, dc.Name)
	if err != nil {
		return nil, err
	}
	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		// Delete ignore labels
		if labelMatches(dnsutil.TrimDomainName(rec.Original.(*cfRecord).Name, dc.Name), c.ignoredLabels) {
			fmt.Printf("ignored_label: %s\n", rec.Original.(*cfRecord).Name)
			records = append(records[:i], records[i+1:]...)
		}
	}
	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" {
			rec.Type = "CNAME"
		}
		if labelMatches(rec.Name, c.ignoredLabels) {
			log.Fatalf("FATAL: dnsconfig contains label that matches ignored_labels: %#v is in %v)\n", rec.Name, c.ignoredLabels)
		}
	}
	checkNSModifications(dc)
	differ := diff.New(dc, getProxyMetadata)
	_, create, del, mod := differ.IncrementalDiff(records)
	corrections := []*models.Correction{}

	for _, d := range del {
		corrections = append(corrections, c.deleteRec(d.Existing.Original.(*cfRecord), id))
	}
	for _, d := range create {
		corrections = append(corrections, c.createRec(d.Desired, id)...)
	}

	for _, d := range mod {
		e, rec := d.Existing.Original.(*cfRecord), d.Desired
		proxy := e.Proxiable && rec.Metadata[metaProxy] != "off"
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return c.modifyRecord(id, e.ID, proxy, rec) },
		})
	}
	return corrections, nil
}

func checkNSModifications(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.NameFQDN == dc.Name {
			if !strings.HasSuffix(rec.Target, ".ns.cloudflare.com.") {
				log.Printf("Warning: cloudflare does not support modifying NS records on base domain. %s will not be added.", rec.Target)
			}
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

const (
	metaProxy         = "cloudflare_proxy"
	metaProxyDefault  = metaProxy + "_default"
	metaOriginalIP    = "original_ip"    // TODO(tlim): Unclear what this means.
	metaIPConversions = "ip_conversions" // TODO(tlim): Rename to obscure_rules.
)

func checkProxyVal(v string) (string, error) {
	v = strings.ToLower(v)
	if v != "on" && v != "off" && v != "full" {
		return "", fmt.Errorf("Bad metadata value for cloudflare_proxy: '%s'. Use on/off/full", v)
	}
	return v, nil
}

func (c *CloudflareApi) preprocessConfig(dc *models.DomainConfig) error {

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

	// Normalize the proxy setting for each record.
	// A and CNAMEs: Validate. If null, set to default.
	// else: Make sure it wasn't set.  Set to default.
	for _, rec := range dc.Records {
		if rec.Metadata == nil {
			rec.Metadata = map[string]string{}
		}
		if rec.TTL == 0 || rec.TTL == 300 {
			rec.TTL = 1
		}
		if rec.TTL != 1 && rec.TTL < 120 {
			rec.TTL = 120
		}
		if rec.Type != "A" && rec.Type != "CNAME" && rec.Type != "AAAA" && rec.Type != "ALIAS" {
			if rec.Metadata[metaProxy] != "" {
				return fmt.Errorf("cloudflare_proxy set on %v record: %#v cloudflare_proxy=%#v", rec.Type, rec.Name, rec.Metadata[metaProxy])
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
	}

	// look for ip conversions and transform records
	for _, rec := range dc.Records {
		if rec.Type != "A" {
			continue
		}
		//only transform "full"
		if rec.Metadata[metaProxy] != "full" {
			continue
		}
		ip := net.ParseIP(rec.Target)
		if ip == nil {
			return fmt.Errorf("%s is not a valid ip address", rec.Target)
		}
		newIP, err := transform.TransformIP(ip, c.ipConversions)
		if err != nil {
			return err
		}
		rec.Metadata[metaOriginalIP] = rec.Target
		rec.Target = newIP.String()
	}

	return nil
}

func newCloudflare(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &CloudflareApi{}
	api.ApiUser, api.ApiKey = m["apiuser"], m["apikey"]
	// check api keys from creds json file
	if api.ApiKey == "" || api.ApiUser == "" {
		return nil, fmt.Errorf("Cloudflare apikey and apiuser must be provided.")
	}

	err := api.fetchDomainList()
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		parsedMeta := &struct {
			IPConversions string   `json:"ip_conversions"`
			IgnoredLabels []string `json:"ignored_labels"`
		}{}
		err := json.Unmarshal([]byte(metadata), parsedMeta)
		if err != nil {
			return nil, err
		}
		// ignored_labels:
		for _, l := range parsedMeta.IgnoredLabels {
			api.ignoredLabels = append(api.ignoredLabels, l)
		}
		// parse provider level metadata
		api.ipConversions, err = transform.DecodeTransformTable(parsedMeta.IPConversions)
		if err != nil {
			return nil, err
		}
	}
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("CLOUDFLAREAPI", newCloudflare, providers.CanUseAlias)
}

// Used on the "existing" records.
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
	Data       interface{} `json:"data"`
	Priority   uint16      `json:"priority"`
}

func (c *cfRecord) toRecord(domain string) *models.RecordConfig {
	//normalize cname,mx,ns records with dots to be consistent with our config format.
	if c.Type == "CNAME" || c.Type == "MX" || c.Type == "NS" {
		c.Content = dnsutil.AddOrigin(c.Content+".", domain)
	}
	return &models.RecordConfig{
		NameFQDN: c.Name,
		Type:     c.Type,
		Target:   c.Content,
		Priority: c.Priority,
		TTL:      c.TTL,
		Original: c,
	}
}

func getProxyMetadata(r *models.RecordConfig) map[string]string {
	if r.Type != "A" && r.Type != "AAAA" && r.Type != "CNAME" {
		return nil
	}
	proxied := false
	if r.Original != nil {
		proxied = r.Original.(*cfRecord).Proxied
	} else {
		proxied = r.Metadata[metaProxy] != "off"
	}
	return map[string]string{
		"proxy": fmt.Sprint(proxied),
	}
}

func (c *CloudflareApi) EnsureDomainExists(domain string) error {
	if _, ok := c.domainIndex[domain]; ok {
		return nil
	}
	var id string
	id, err := c.createZone(domain)
	fmt.Printf("Added zone for %s to Cloudflare account: %s\n", domain, id)
	return err
}
