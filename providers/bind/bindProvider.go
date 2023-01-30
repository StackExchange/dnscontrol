package bind

/*

bind -
  Generate zonefiles suitable for BIND.

	The zonefiles are read and written to the directory -bind_dir

	If the old zonefiles are readable, we read them to determine
	if an update is actually needed. The old zonefile is also used
	as the basis for generating the new SOA serial number.

*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/miekg/dns"
)

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Can("Just writes out a comment indicating DNSSEC was requested"),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSOA:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Can("Driver just maintains list of zone files. It should automatically add missing ones."),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
}

func initBind(config map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	// config -- the key/values from creds.json
	// meta -- the json blob from NewReq('name', 'TYPE', meta)
	api := &bindProvider{
		directory:      config["directory"],
		filenameformat: config["filenameformat"],
	}
	if api.directory == "" {
		api.directory = "zones"
	}
	if api.filenameformat == "" {
		api.filenameformat = "%U.zone"
	}
	if len(providermeta) != 0 {
		err := json.Unmarshal(providermeta, api)
		if err != nil {
			return nil, err
		}
	}
	var nss []string
	for i, ns := range api.DefaultNS {
		if ns == "" {
			return nil, fmt.Errorf("empty string in default_ns[%d]", i)
		}
		// If it contains a ".", it must end in a ".".
		if strings.ContainsRune(ns, '.') && ns[len(ns)-1] != '.' {
			return nil, fmt.Errorf("default_ns (%v) must end with a (.) [https://docs.dnscontrol.org/language-reference/why-the-dot]", ns)
		}
		// This is one of the (increasingly rare) cases where we store a
		// name without the trailing dot to indicate a FQDN.
		nss = append(nss, strings.TrimSuffix(ns, "."))
	}
	var err error
	api.nameservers, err = models.ToNameservers(nss)
	return api, err
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   initBind,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("BIND", fns, features)
}

// SoaDefaults contains the parts of the default SOA settings.
type SoaDefaults struct {
	Ns      string `json:"master"`
	Mbox    string `json:"mbox"`
	Serial  uint32 `json:"serial"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	Minttl  uint32 `json:"minttl"`
	TTL     uint32 `json:"ttl,omitempty"`
}

func (s SoaDefaults) String() string {
	return fmt.Sprintf("%s %s %d %d %d %d %d %d", s.Ns, s.Mbox, s.Serial, s.Refresh, s.Retry, s.Expire, s.Minttl, s.TTL)
}

// bindProvider is the provider handle for the bindProvider driver.
type bindProvider struct {
	DefaultNS      []string    `json:"default_ns"`
	DefaultSoa     SoaDefaults `json:"default_soa"`
	nameservers    []*models.Nameserver
	directory      string
	filenameformat string
	zonefile       string // Where the zone data is expected
	zoneFileFound  bool   // Did the zonefile exist?
}

// GetNameservers returns the nameservers for a domain.
func (c *bindProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	var r []string
	for _, j := range c.nameservers {
		r = append(r, j.Name)
	}
	return models.ToNameservers(r)
}

// ListZones returns all the zones in an account
func (c *bindProvider) ListZones() ([]string, error) {
	if _, err := os.Stat(c.directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory %q does not exist", c.directory)
	}

	var files []string
	f, err := os.Open(c.directory)
	if err != nil {
		return files, fmt.Errorf("bind ListZones open dir %q: %w",
			c.directory, err)
	}
	filenames, err := f.Readdirnames(-1)
	if err != nil {
		return files, fmt.Errorf("bind ListZones readdir %q: %w",
			c.directory, err)
	}

	return extractZonesFromFilenames(c.filenameformat, filenames), nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *bindProvider) GetZoneRecords(domain string) (models.Records, error) {
	foundRecords := models.Records{}

	if _, err := os.Stat(c.directory); os.IsNotExist(err) {
		printer.Printf("\nWARNING: BIND directory %q does not exist!\n", c.directory)
	}

	if c.zonefile == "" {
		// This layering violation is needed for tests only.
		// Otherwise, this is set already.
		c.zonefile = filepath.Join(c.directory,
			makeFileName(c.filenameformat, domain, domain, ""))
	}
	content, err := os.ReadFile(c.zonefile)
	if os.IsNotExist(err) {
		// If the file doesn't exist, that's not an error. Just informational.
		c.zoneFileFound = false
		fmt.Fprintf(os.Stderr, "File does not yet exist: %q (will create)\n", c.zonefile)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", c.zonefile, err)
	}
	c.zoneFileFound = true

	zp := dns.NewZoneParser(strings.NewReader(string(content)), domain, c.zonefile)

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rec, err := models.RRtoRC(rr, domain)
		if err != nil {
			return nil, err
		}
		foundRecords = append(foundRecords, &rec)
	}

	if err := zp.Err(); err != nil {
		return nil, fmt.Errorf("error while parsing '%v': %w", c.zonefile, err)
	}
	return foundRecords, nil
}

// GetDomainCorrections returns a list of corrections to update a domain.
func (c *bindProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	comments := make([]string, 0, 5)
	comments = append(comments,
		fmt.Sprintf("generated with dnscontrol %s", time.Now().Format(time.RFC3339)),
	)
	if dc.AutoDNSSEC == "on" {
		// This does nothing but reminds the user to add the correct
		// auto-dnssecc zone statement to named.conf.
		// While it is a no-op, it is useful for situations where a zone
		// has multiple providers.
		comments = append(comments, "Automatic DNSSEC signing requested")
	}

	c.zonefile = filepath.Join(c.directory,
		makeFileName(c.filenameformat, dc.UniqueName, dc.Name, dc.Tag))

	foundRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Find the SOA records; use them to make or update the desired SOA.
	var foundSoa *models.RecordConfig
	for _, r := range foundRecords {
		if r.Type == "SOA" && r.Name == "@" {
			foundSoa = r
			break
		}
	}
	var desiredSoa *models.RecordConfig
	for _, r := range dc.Records {
		if r.Type == "SOA" && r.Name == "@" {
			desiredSoa = r
			break
		}
	}
	soaRec, nextSerial := makeSoa(dc.Name, &c.DefaultSoa, foundSoa, desiredSoa)
	if desiredSoa == nil {
		dc.Records = append(dc.Records, soaRec)
		desiredSoa = dc.Records[len(dc.Records)-1]
	} else {
		*desiredSoa = *soaRec
	}

	// Normalize
	models.PostProcessRecords(foundRecords)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	changes := false
	var msg string

	if !diff2.EnableDiff2 {

		differ := diff.New(dc)
		_, create, del, mod, err := differ.IncrementalDiff(foundRecords)
		if err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}
		// Print a list of changes. Generate an actual change that is the zone

		for _, i := range create {
			changes = true
			if c.zoneFileFound {
				fmt.Fprintln(buf, i)
			}
		}
		for _, i := range del {
			changes = true
			if c.zoneFileFound {
				fmt.Fprintln(buf, i)
			}
		}
		for _, i := range mod {
			changes = true
			if c.zoneFileFound {
				fmt.Fprintln(buf, i)
			}
		}

		if c.zoneFileFound {
			msg = fmt.Sprintf("GENERATE_ZONEFILE: '%s'. Changes:\n%s", dc.Name, buf)
		} else {
			msg = fmt.Sprintf("GENERATE_ZONEFILE: '%s' (new file with %d records)\n", dc.Name, len(create))
		}

	} else {

		var msgs []string
		msgs, changes, err = diff2.ByZone(foundRecords, dc, nil)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("DEBUG: BIND changes=%v\n", changes)
		msg = strings.Join(msgs, "\n")

	}

	var corrections []*models.Correction
	//fmt.Printf("DEBUG: BIND changes=%v\n", changes)
	if changes {

		// We only change the serial number if there is a change.
		desiredSoa.SoaSerial = nextSerial

		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					printer.Printf("WRITING ZONEFILE: %v\n", c.zonefile)
					zf, err := os.Create(c.zonefile)
					if err != nil {
						return fmt.Errorf("could not create zonefile: %w", err)
					}
					// Beware that if there are any fake types, then they will
					// be commented out on write, but we don't reverse that when
					// reading, so there will be a diff on every invocation.
					err = prettyzone.WriteZoneFileRC(zf, dc.Records, dc.Name, 0, comments)

					if err != nil {
						return fmt.Errorf("failed WriteZoneFile: %w", err)
					}
					err = zf.Close()
					if err != nil {
						return fmt.Errorf("closing: %w", err)
					}
					return nil
				},
			})
	}

	return corrections, nil
}
