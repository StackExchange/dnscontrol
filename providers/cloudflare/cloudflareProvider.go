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

Cloudflare APi DNS provider:

Info required in `creds.json`:
   - apikey
   - apiuser

Record level metadata availible:
   - cloudflare_proxy ("true" or "false")

Domain level metadata availible:
   - cloudflare_proxy_default ("true" or "false")

 Provider level metadata availible:
   - ip_conversions
   - secret_ips
*/

type CloudflareApi struct {
	ApiKey        string `json:"apikey"`
	ApiUser       string `json:"apiuser"`
	domainIndex   map[string]string
	nameservers   map[string][]string
	ipConversions []transform.IpConversion
	secretIPs     []net.IP
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
func (c *CloudflareApi) GetNameservers(domain string) ([]string, error) {
	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	ns, ok := c.nameservers[domain]
	if !ok {
		return nil, fmt.Errorf("Nameservers for %s not found in cloudflare account", domain)
	}
	return ns, nil
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
	records, err := c.getRecordsForDomain(id)
	if err != nil {
		return nil, err
	}
	//for _, rec := range records {
	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		// Delete ignore labels
		if labelMatches(dnsutil.TrimDomainName(rec.(*cfRecord).Name, dc.Name), c.ignoredLabels) {
			fmt.Printf("ignored_label: %s\n", rec.(*cfRecord).Name)
			records = append(records[:i], records[i+1:]...)
		}
		//normalize cname,mx,ns records with dots to be consistent with our config format.
		t := rec.(*cfRecord).Type
		if t == "CNAME" || t == "MX" || t == "NS" {
			rec.(*cfRecord).Content = dnsutil.AddOrigin(rec.(*cfRecord).Content+".", dc.Name)
		}
	}

	expectedRecords := make([]diff.Record, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if labelMatches(rec.Name, c.ignoredLabels) {
			log.Fatalf("FATAL: dnsconfig contains label that matches ignored_labels: %#v is in %v)\n", rec.Name, c.ignoredLabels)
			// Since we log.Fatalf, we don't need to be clean here.
		}
		expectedRecords = append(expectedRecords, recordWrapper{rec})
	}
	_, create, del, mod := diff.IncrementalDiff(records, expectedRecords)
	corrections := []*models.Correction{}

	for _, d := range del {
		corrections = append(corrections, c.deleteRec(d.Existing.(*cfRecord), id))
	}
	for _, d := range create {
		corrections = append(corrections, c.createRec(d.Desired.(recordWrapper).RecordConfig, id)...)
	}

	for _, d := range mod {
		e, rec := d.Existing.(*cfRecord), d.Desired.(recordWrapper)
		proxy := e.Proxiable && rec.Metadata[metaProxy] != "off"
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("MODIFY record %s %s: (%s %s) => (%s %s)", rec.Name, rec.Type, e.Content, e.GetComparisionData(), rec.Target, rec.GetComparisionData()),
			F:   func() error { return c.modifyRecord(id, e.ID, proxy, rec.RecordConfig) },
		})
	}
	return corrections, nil
}

const (
	metaProxy         = "cloudflare_proxy"
	metaProxyDefault  = metaProxy + "_default"
	metaOriginalIP    = "original_ip"    // TODO(tlim): Unclear what this means.
	metaIPConversions = "ip_conversions" // TODO(tlim): Rename to obscure_rules.
	metaSecretIPs     = "secret_ips"     // TODO(tlim): Rename to obscured_cidrs.
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
		if rec.Type != "A" && rec.Type != "CNAME" && rec.Type != "AAAA" {
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
		if rec.TTL == 0 {
			rec.TTL = 1
		}
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

	if len(metadata) > 0 {
		parsedMeta := &struct {
			IPConversions string        `json:"ip_conversions"`
			SecretIps     []interface{} `json:"secret_ips"`
			IgnoredLabels []string      `json:"ignored_labels"`
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
		ips := []net.IP{}
		for _, ipStr := range parsedMeta.SecretIps {
			var ip net.IP
			if ip, err = models.InterfaceToIP(ipStr); err != nil {
				return nil, err
			}
			ips = append(ips, ip)
		}
		api.secretIPs = ips
	}
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("CLOUDFLAREAPI", newCloudflare)
}

// Used on the "existing" records.
type cfRecord struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Content    string      `json:"content"`
	Proxiable  bool        `json:"proxiable"`
	Proxied    bool        `json:"proxied"`
	TTL        int         `json:"ttl"`
	Locked     bool        `json:"locked"`
	ZoneID     string      `json:"zone_id"`
	ZoneName   string      `json:"zone_name"`
	CreatedOn  time.Time   `json:"created_on"`
	ModifiedOn time.Time   `json:"modified_on"`
	Data       interface{} `json:"data"`
	Priority   int         `json:"priority"`
}

func (c *cfRecord) GetName() string {
	return c.Name
}

func (c *cfRecord) GetType() string {
	return c.Type
}

func (c *cfRecord) GetContent() string {
	return c.Content
}

func (c *cfRecord) GetComparisionData() string {
	mxPrio := ""
	if c.Type == "MX" {
		mxPrio = fmt.Sprintf(" %d ", c.Priority)
	}
	proxy := ""
	if c.Type == "A" || c.Type == "CNAME" || c.Type == "AAAA" {
		proxy = fmt.Sprintf(" proxy=%v ", c.Proxied)
	}
	return fmt.Sprintf("%d%s%s", c.TTL, mxPrio, proxy)
}

// Used on the "expected" records.
type recordWrapper struct {
	*models.RecordConfig
}

func (c recordWrapper) GetComparisionData() string {
	mxPrio := ""
	if c.Type == "MX" {
		mxPrio = fmt.Sprintf(" %d ", c.Priority)
	}
	proxy := ""
	if c.Type == "A" || c.Type == "AAAA" || c.Type == "CNAME" {
		proxy = fmt.Sprintf(" proxy=%v ", c.Metadata[metaProxy] != "off")
	}

	ttl := c.TTL
	if ttl == 0 {
		ttl = 1
	}
	return fmt.Sprintf("%d%s%s", ttl, mxPrio, proxy)
}
