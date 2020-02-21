package bind

/*

bind -
  Generate zonefiles suitiable for BIND.

	The zonefiles are read and written to the directory -bind_dir

	If the old zonefiles are readable, we read them to determine
	if an update is actually needed. The old zonefile is also used
	as the basis for generating the new SOA serial number.

*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/prettyzone"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/diff"
)

var features = providers.DocumentationNotes{
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Can("Driver just maintains list of zone files. It should automatically add missing ones."),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanGetZones:            providers.Can(),
}

func initBind(config map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	// config -- the key/values from creds.json
	// meta -- the json blob from NewReq('name', 'TYPE', meta)
	api := &Bind{
		directory: config["directory"],
	}
	if api.directory == "" {
		api.directory = "zones"
	}
	if len(providermeta) != 0 {
		err := json.Unmarshal(providermeta, api)
		if err != nil {
			return nil, err
		}
	}
	api.nameservers = models.StringsToNameservers(api.DefaultNS)
	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("BIND", initBind, features)
}

// SoaInfo contains the parts of a SOA rtype.
type SoaInfo struct {
	Ns      string `json:"master"`
	Mbox    string `json:"mbox"`
	Serial  uint32 `json:"serial"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	Minttl  uint32 `json:"minttl"`
}

func (s SoaInfo) String() string {
	return fmt.Sprintf("%s %s %d %d %d %d %d", s.Ns, s.Mbox, s.Serial, s.Refresh, s.Retry, s.Expire, s.Minttl)
}

// Bind is the provider handle for the Bind driver.
type Bind struct {
	DefaultNS     []string `json:"default_ns"`
	DefaultSoa    SoaInfo  `json:"default_soa"`
	nameservers   []*models.Nameserver
	directory     string
	zonefile      string // Where the zone data is expected
	zoneFileFound bool   // Did the zonefile exist?
}

func makeDefaultSOA(info SoaInfo, origin string) *models.RecordConfig {
	// Make a default SOA record in case one isn't found:
	soaRec := models.RecordConfig{
		Type: "SOA",
	}
	soaRec.SetLabel("@", origin)
	if len(info.Ns) == 0 {
		info.Ns = "DEFAULT_NOT_SET."
	}
	if len(info.Mbox) == 0 {
		info.Mbox = "DEFAULT_NOT_SET."
	}
	if info.Serial == 0 {
		info.Serial = 1
	}
	if info.Refresh == 0 {
		info.Refresh = 3600
	}
	if info.Retry == 0 {
		info.Retry = 600
	}
	if info.Expire == 0 {
		info.Expire = 604800
	}
	if info.Minttl == 0 {
		info.Minttl = 1440
	}
	soaRec.SetTarget(info.String())

	return &soaRec
}

// GetNameservers returns the nameservers for a domain.
func (c *Bind) GetNameservers(string) ([]*models.Nameserver, error) {
	return c.nameservers, nil
}

// ListZones returns all the zones in an account
func (c *Bind) ListZones() ([]string, error) {
	if _, err := os.Stat(c.directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("BIND directory %q does not exist!\n", c.directory)
	}

	filenames, err := filepath.Glob(filepath.Join(c.directory, "*.zone"))
	if err != nil {
		return nil, err
	}
	var zones []string
	for _, n := range filenames {
		_, file := filepath.Split(n)
		zones = append(zones, strings.TrimSuffix(file, ".zone"))
	}
	return zones, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *Bind) GetZoneRecords(domain string) (models.Records, error) {

	// Default SOA record.  If we see one in the zone, this will be replaced.
	soaRec := makeDefaultSOA(c.DefaultSoa, domain)
	foundRecords := models.Records{}
	var oldSerial, newSerial uint32

	if _, err := os.Stat(c.directory); os.IsNotExist(err) {
		fmt.Printf("\nWARNING: BIND directory %q does not exist!\n", c.directory)
	}

	zonefile := filepath.Join(c.directory, strings.Replace(strings.ToLower(domain), "/", "_", -1)+".zone")
	c.zonefile = zonefile
	foundFH, err := os.Open(zonefile)
	c.zoneFileFound = err == nil
	if err != nil && !os.IsNotExist(os.ErrNotExist) {
		// Don't whine if the file doesn't exist. However all other
		// errors will be reported.
		fmt.Printf("Could not read zonefile: %v\n", err)
	} else {
		for x := range dns.ParseZone(foundFH, domain, zonefile) {
			if x.Error != nil {
				log.Println("Error in zonefile:", x.Error)
			} else {
				rec, serial := models.RRtoRC(x.RR, domain, oldSerial)
				if serial != 0 && oldSerial != 0 {
					log.Fatalf("Multiple SOA records in zonefile: %v\n", zonefile)
				}
				if serial != 0 {
					// This was an SOA record. Update the serial.
					oldSerial = serial
					newSerial = generateSerial(oldSerial)
					// Regenerate with new serial:
					*soaRec, _ = models.RRtoRC(x.RR, domain, newSerial)
					rec = *soaRec
				}
				foundRecords = append(foundRecords, &rec)
			}
		}
	}

	// Add SOA record to expected set:
	if !foundRecords.HasRecordTypeName("SOA", "@") {
		//foundRecords = append(foundRecords, soaRec)
	}

	return foundRecords, nil
}

// GetDomainCorrections returns a list of corrections to update a domain.
func (c *Bind) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	// Phase 1: Copy everything to []*models.RecordConfig:
	//    expectedRecords < dc.Records[i]
	//    foundRecords < zonefile
	//
	// Phase 2: Do any manipulations:
	// add NS
	// manipulate SOA
	//
	// Phase 3: Convert to []diff.Records and compare:
	// expectedDiffRecords < expectedRecords
	// foundDiffRecords < foundRecords
	// diff.Inc...(foundDiffRecords, expectedDiffRecords )

	foundRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(foundRecords)

	buf := &bytes.Buffer{}
	// Print a list of changes. Generate an actual change that is the zone
	changes := false
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
	msg := fmt.Sprintf("GENERATE_ZONEFILE: %s\n", dc.Name)
	if !c.zoneFileFound {
		msg = msg + fmt.Sprintf(" (%d records)\n", len(create))
	}
	msg += buf.String()
	corrections := []*models.Correction{}
	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					fmt.Printf("CREATING ZONEFILE: %v\n", c.zonefile)
					zf, err := os.Create(c.zonefile)
					if err != nil {
						log.Fatalf("Could not create zonefile: %v", err)
					}
					err = prettyzone.WriteZoneFileRC(zf, dc.Records, dc.Name)

					if err != nil {
						log.Fatalf("WriteZoneFile error: %v\n", err)
					}
					err = zf.Close()
					if err != nil {
						log.Fatalf("Closing: %v", err)
					}
					return nil
				},
			})
	}

	return corrections, nil
}
