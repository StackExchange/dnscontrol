package axfrddns

/*

axfrddns -
  Fetch the zone with AXFR request to a given primary master, and
  push DynamicDNS updates to the same server.

  Both the AXFR request and the updates might be authentified with a
  TSIG.

*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

const (
	// Be very conservative..
	// TODO add a configuration ??
	dnsTimeout         time.Duration = 30 * time.Second
	dnssec_dummy_label               = "__dnssec"
	dnssec_dummy_txt                 = "Domain has DNSSec records, not displayed here."
)

var features = providers.DocumentationNotes{
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanAutoDNSSEC:          providers.Can("Just warn when DNSSEC is requested but no RRSIG is found in the AXFR or warn when DNSSEC is not requested but RRSIG are found in the AXFR."),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

func initAxfrDdns(config map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	// config -- the key/values from creds.json
	// meta -- the json blob from NewReq('name', 'TYPE', meta)
	var err error
	api := &AxfrDdns{
		rand: rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
	}
	param := &Param{}
	if len(providermeta) != 0 {
		err := json.Unmarshal(providermeta, param)
		if err != nil {
			return nil, err
		}
	}
	var nss []string
	if config["nameservers"] != "" {
		nss = strings.Split(config["nameservers"], ",")
	}
	for _, ns := range param.DefaultNS {
		nss = append(nss, ns[0:len(ns)-1])
	}
	api.nameservers, err = models.ToNameservers(nss)
	if err != nil {
		return nil, err
	}
	if config["master"] != "" {
		api.master = config["master"]
		if !strings.Contains(api.master, ":") {
			api.master = api.master + ":53"
		}
	} else if len(api.nameservers) != 0 {
		api.master = api.nameservers[0].Name + ":53"
	} else {
		return nil, fmt.Errorf("TODO: nameservers cannot be null")
	}
	api.updateKey, err = readKey(config["update-key"])
	if err != nil {
		return nil, err
	}
	api.transferKey, err = readKey(config["transfer-key"])
	if err != nil {
		return nil, err
	}
	// TODO check for unexpected key in config

	return api, err
}

func init() {
	providers.RegisterDomainServiceProviderType("AXFRDDNS", initAxfrDdns, features)
}

func readKey(raw string) (*Key, error) {
	if raw == "" {
		return nil, nil
	}
	arr := strings.Split(raw, ":")
	if len(arr) != 3 {
		return nil, fmt.Errorf("TODO: invalid key format")
	}
	var algo string
	switch arr[0] {
	case "hmac-md5", "md5":
		algo = dns.HmacMD5
	case "hmac-sha1", "sha1":
		algo = dns.HmacSHA1
	case "hmac-sha256", "sha256":
		algo = dns.HmacSHA256
	case "hmac-sha512", "sha512":
		algo = dns.HmacSHA512
	default:
		return nil, fmt.Errorf("TODO: unknown algo")
	}
	// TODO ensure that secret is valid base64
	return &Key{algo: algo, id: arr[1] + ".", secret: arr[2]}, nil
}

type Param struct {
	DefaultNS []string `json:"default_ns"`
}

type Key struct {
	algo   string
	id     string
	secret string
}

type AxfrDdns struct {
	rand        *rand.Rand
	master      string
	nameservers []*models.Nameserver
	transferKey *Key
	updateKey   *Key
}

// GetNameservers returns the nameservers for a domain.
func (c *AxfrDdns) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return c.nameservers, nil
}

// FetchZoneRecords gets the records of a zone and returns them in dns.RR format.
func (c *AxfrDdns) FetchZoneRecords(domain string) ([]dns.RR, error) {

	transfer := new(dns.Transfer)
	transfer.DialTimeout = dnsTimeout
	transfer.ReadTimeout = dnsTimeout

	request := new(dns.Msg)
	request.SetAxfr(domain + ".")

	if c.transferKey != nil {
		transfer.TsigSecret =
			map[string]string{c.transferKey.id: c.transferKey.secret}
		request.SetTsig(c.transferKey.id, c.transferKey.algo, 300, time.Now().Unix())
	}

	envelope, err := transfer.In(request, c.master)
	if err != nil {
		return nil, err
	}

	var rawRecords []dns.RR
	for msg := range envelope {
		if msg.Error != nil {
			// Fragile but more "user-friendly" error-handling
			if msg.Error.Error() == "dns: bad xfr rcode: 9" {
				return nil, fmt.Errorf("dns: NOT AUTH")
			}
			return nil, msg.Error
		}
		rawRecords = append(rawRecords, msg.RR...)
	}
	return rawRecords, nil

}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *AxfrDdns) GetZoneRecords(domain string) (models.Records, error) {

	rawRecords, err := c.FetchZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	var foundDNSSecRecords *models.RecordConfig
	foundRecords := models.Records{}
	for _, rr := range rawRecords {
		switch rr.(type) {
		case *dns.RRSIG,
			*dns.DNSKEY,
			*dns.CDNSKEY,
			*dns.CDS,
			*dns.NSEC,
			*dns.NSEC3,
			*dns.NSEC3PARAM:
			// Ignoring DNSSec RRs, but replacing it with a single
			// "TXT" placeholder
			if foundDNSSecRecords == nil {
				foundDNSSecRecords = new(models.RecordConfig)
				foundDNSSecRecords.Type = "TXT"
				foundDNSSecRecords.SetLabel(dnssec_dummy_label, domain)
				err = foundDNSSecRecords.SetTargetTXT(dnssec_dummy_txt)
				if err != nil {
					return nil, err
				}
			}
			continue
		default:
			rec := models.RRtoRC(rr, domain)
			foundRecords = append(foundRecords, &rec)
		}
	}

	if len(foundRecords) >= 1 && foundRecords[len(foundRecords)-1].Type == "SOA" {
		// When TSig is used, the SOA is sent two times: as the
		// first and the last record
		foundRecords = foundRecords[:len(foundRecords)-1]
	}

	if foundDNSSecRecords != nil {
		foundRecords = append(foundRecords, foundDNSSecRecords)
	}

	return foundRecords, nil

}

// GetDomainCorrections returns a list of corrections to update a domain.
func (c *AxfrDdns) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	foundRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	if len(foundRecords) >= 1 && foundRecords[0].Type == "SOA" {
		// Ignoring the SOA, others providers  don't manage it either.
		foundRecords = foundRecords[1:]
	}

	hasDnssecRecords := false
	if len(foundRecords) >= 1 {
		last := foundRecords[len(foundRecords)-1]
		if last.Type == "TXT" &&
			last.Name == dnssec_dummy_label &&
			len(last.TxtStrings) == 1 &&
			last.TxtStrings[0] == dnssec_dummy_txt {
			hasDnssecRecords = true
			foundRecords = foundRecords[0:(len(foundRecords) - 1)]
		}
	}

	if dc.AutoDNSSEC && !hasDnssecRecords {
		fmt.Printf("Warning: AUTODNSSEC is set, but no DNSKEY or RRSIG record was found in the AXFR answer!\n")
	} else if !dc.AutoDNSSEC && hasDnssecRecords {
		fmt.Printf("Warning: AUTODNSSEC is not set, but DNSKEY or RRSIG records were found in the AXFR answer!\n")
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
		fmt.Fprintln(buf, i)
	}
	for _, i := range del {
		changes = true
		fmt.Fprintln(buf, i)
	}
	for _, i := range mod {
		changes = true
		fmt.Fprintln(buf, i)
	}
	msg := fmt.Sprintf("DDNS UPDATES to '%s' (primary master: '%s'). Changes:\n%s", dc.Name, c.master, buf)

	corrections := []*models.Correction{}
	if changes {

		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {

					update := new(dns.Msg)
					update.SetUpdate(dc.Name + ".")
					update.Id = uint16(c.rand.Intn(math.MaxUint16))
					for _, c := range del {
						update.Remove([]dns.RR{c.Existing.ToRR()})
					}
					for _, c := range mod {
						update.Remove([]dns.RR{c.Existing.ToRR()})
						update.Insert([]dns.RR{c.Desired.ToRR()})
					}
					for _, c := range create {
						update.Insert([]dns.RR{c.Desired.ToRR()})
					}

					client := new(dns.Client)
					client.Timeout = dnsTimeout
					if c.updateKey != nil {
						client.TsigSecret =
							map[string]string{c.updateKey.id: c.updateKey.secret}
						update.SetTsig(c.updateKey.id, c.updateKey.algo, 300, time.Now().Unix())
					}

					msg, _, err := client.Exchange(update, c.master)
					if err != nil {
						return err
					}
					if msg.MsgHdr.Rcode != 0 {
						return fmt.Errorf("Update failed: %s (%d).",
							dns.RcodeToString[msg.MsgHdr.Rcode],
							msg.MsgHdr.Rcode)
					}

					return nil
				},
			})
	}
	return corrections, nil
}
