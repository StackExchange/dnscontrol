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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
)

func initBind(config map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	// config -- the key/values from creds.json
	// meta -- the json blob from NewReq('name', 'TYPE', meta)

	api := &Bind{}
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
	providers.RegisterDomainServiceProviderType("BIND", initBind, providers.CanUsePTR, providers.CanUseSRV, providers.CanUseCAA)
}

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

type Bind struct {
	DefaultNS   []string `json:"default_ns"`
	DefaultSoa  SoaInfo  `json:"default_soa"`
	nameservers []*models.Nameserver
}

var bindBaseDir = flag.String("bindtree", "zones", "BIND: Directory that stores BIND zonefiles.")

//var bindSkeletin = flag.String("bind_skeletin", "skeletin/master/var/named/chroot/var/named/master", "")

func rrToRecord(rr dns.RR, origin string, replaceSerial uint32) (models.RecordConfig, uint32) {
	// Convert's dns.RR into our native data type (models.RecordConfig).
	// Records are translated directly with no changes.
	// If it is an SOA for the apex domain and
	// replaceSerial != 0, change the serial to replaceSerial.
	// WARNING(tlim): This assumes SOAs do not have serial=0.
	// If one is found, we replace it with serial=1.
	var old_serial, new_serial uint32
	header := rr.Header()
	rc := models.RecordConfig{}
	rc.Type = dns.TypeToString[header.Rrtype]
	rc.NameFQDN = strings.ToLower(strings.TrimSuffix(header.Name, "."))
	rc.Name = strings.ToLower(dnsutil.TrimDomainName(header.Name, origin))
	rc.TTL = header.Ttl
	switch v := rr.(type) {
	case *dns.A:
		rc.Target = v.A.String()
	case *dns.AAAA:
		rc.Target = v.AAAA.String()
	case *dns.CAA:
		rc.CaaTag = v.Tag
		rc.CaaFlag = v.Flag
		rc.Target = v.Value
	case *dns.CNAME:
		rc.Target = v.Target
	case *dns.MX:
		rc.Target = v.Mx
		rc.MxPreference = v.Preference
	case *dns.NS:
		rc.Target = v.Ns
	case *dns.PTR:
		rc.Target = v.Ptr
	case *dns.SOA:
		old_serial = v.Serial
		if old_serial == 0 {
			// For SOA records, we never return a 0 serial number.
			old_serial = 1
		}
		new_serial = v.Serial
		if rc.Name == "@" && replaceSerial != 0 {
			new_serial = replaceSerial
		}
		rc.Target = fmt.Sprintf("%v %v %v %v %v %v %v",
			v.Ns, v.Mbox, new_serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
	case *dns.SRV:
		rc.Target = v.Target
		rc.SrvPort = v.Port
		rc.SrvWeight = v.Weight
		rc.SrvPriority = v.Priority
	case *dns.TXT:
		rc.Target = strings.Join(v.Txt, " ")
	default:
		log.Fatalf("rrToRecord: Unimplemented zone record type=%s (%v)\n", rc.Type, rr)
	}
	return rc, old_serial
}

func makeDefaultSOA(info SoaInfo, origin string) *models.RecordConfig {
	// Make a default SOA record in case one isn't found:
	soaRec := models.RecordConfig{
		Type: "SOA",
		Name: "@",
	}
	soaRec.NameFQDN = dnsutil.AddOrigin(soaRec.Name, origin)
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
	soaRec.Target = info.String()

	return &soaRec
}

func (c *Bind) GetNameservers(string) ([]*models.Nameserver, error) {
	return c.nameservers, nil
}

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

	// Default SOA record.  If we see one in the zone, this will be replaced.
	soaRec := makeDefaultSOA(c.DefaultSoa, dc.Name)

	// Read foundRecords:
	foundRecords := make([]*models.RecordConfig, 0)
	var oldSerial, newSerial uint32
	zonefile := filepath.Join(*bindBaseDir, strings.Replace(strings.ToLower(dc.Name), "/", "_", -1)+".zone")
	foundFH, err := os.Open(zonefile)
	zoneFileFound := err == nil
	if err != nil && !os.IsNotExist(os.ErrNotExist) {
		// Don't whine if the file doesn't exist. However all other
		// errors will be reported.
		fmt.Printf("Could not read zonefile: %v\n", err)
	} else {
		for x := range dns.ParseZone(foundFH, dc.Name, zonefile) {
			if x.Error != nil {
				log.Println("Error in zonefile:", x.Error)
			} else {
				rec, serial := rrToRecord(x.RR, dc.Name, oldSerial)
				if serial != 0 && oldSerial != 0 {
					log.Fatalf("Multiple SOA records in zonefile: %v\n", zonefile)
				}
				if serial != 0 {
					// This was an SOA record. Update the serial.
					oldSerial = serial
					newSerial = generate_serial(oldSerial)
					// Regenerate with new serial:
					*soaRec, _ = rrToRecord(x.RR, dc.Name, newSerial)
					rec = *soaRec
				}
				foundRecords = append(foundRecords, &rec)
			}
		}
	}

	// Add SOA record to expected set:
	if !dc.HasRecordTypeName("SOA", "@") {
		dc.Records = append(dc.Records, soaRec)
	}

	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(foundRecords)

	buf := &bytes.Buffer{}
	// Print a list of changes. Generate an actual change that is the zone
	changes := false
	for _, i := range create {
		changes = true
		if zoneFileFound {
			fmt.Fprintln(buf, i)
		}
	}
	for _, i := range del {
		changes = true
		if zoneFileFound {
			fmt.Fprintln(buf, i)
		}
	}
	for _, i := range mod {
		changes = true
		if zoneFileFound {
			fmt.Fprintln(buf, i)
		}
	}
	msg := fmt.Sprintf("GENERATE_ZONEFILE: %s\n", dc.Name)
	if !zoneFileFound {
		msg = msg + fmt.Sprintf(" (%d records)\n", len(create))
	}
	msg += buf.String()
	corrections := []*models.Correction{}
	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					fmt.Printf("CREATING ZONEFILE: %v\n", zonefile)
					zf, err := os.Create(zonefile)
					if err != nil {
						log.Fatalf("Could not create zonefile: %v", err)
					}
					zonefilerecords := make([]dns.RR, 0, len(dc.Records))
					for _, r := range dc.Records {
						zonefilerecords = append(zonefilerecords, r.ToRR())
					}
					err = WriteZoneFile(zf, zonefilerecords, dc.Name)

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
