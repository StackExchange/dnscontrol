package axfrddns

/*

axfrddns -
  Fetch the zone with an AXFR request (RFC5936) to a given primary master, and
  push Dynamic DNS updates (RFC2136) to the same server.

  Both the AXFR request and the updates might be authentificated with
  a TSIG.

*/

import (
	"bytes"
	"encoding/base64"
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
	dnsTimeout       = 30 * time.Second
	dnssecDummyLabel = "__dnssec"
	dnssecDummyTxt   = "Domain has DNSSec records, not displayed here."
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

// axfrddnsProvider stores the client info for the provider.
type axfrddnsProvider struct {
	rand        *rand.Rand
	master      string
	nameservers []*models.Nameserver
	transferKey *Key
	updateKey   *Key
}

func initAxfrDdns(config map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	// config -- the key/values from creds.json
	// providermeta -- the json blob from NewReq('name', 'TYPE', providermeta)
	var err error
	api := &axfrddnsProvider{
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
		return nil, fmt.Errorf("nameservers list is empty: creds.json needs a default `nameservers` or an explicit `master`")
	}
	api.updateKey, err = readKey(config["update-key"], "update-key")
	if err != nil {
		return nil, err
	}
	api.transferKey, err = readKey(config["transfer-key"], "transfer-key")
	if err != nil {
		return nil, err
	}
	for key := range config {
		switch key {
		case "master",
			"nameservers",
			"update-key",
			"transfer-key":
			continue
		default:
			fmt.Printf("[Warning] AXFRDDNS: unknown key in `creds.json` (%s)\n", key)
		}
	}
	return api, err
}

func init() {
	providers.RegisterDomainServiceProviderType("AXFRDDNS", initAxfrDdns, features)
}

// Param is used to decode extra parameters sent to provider.
type Param struct {
	DefaultNS []string `json:"default_ns"`
}

// Key stores the individual parts of a TSIG key.
type Key struct {
	algo   string
	id     string
	secret string
}

func readKey(raw string, kind string) (*Key, error) {
	if raw == "" {
		return nil, nil
	}
	arr := strings.Split(raw, ":")
	if len(arr) != 3 {
		return nil, fmt.Errorf("invalid key format (%s) in AXFRDDNS.TSIG", kind)
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
		return nil, fmt.Errorf("unknown algorithm (%s) in AXFRDDNS.TSIG", kind)
	}
	_, err := base64.StdEncoding.DecodeString(arr[2])
	if err != nil {
		return nil, fmt.Errorf("cannot decode Base64 secret (%s) in AXFRDDNS.TSIG", kind)
	}
	return &Key{algo: algo, id: arr[1] + ".", secret: arr[2]}, nil
}

// GetNameservers returns the nameservers for a domain.
func (c *axfrddnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return c.nameservers, nil
}

// FetchZoneRecords gets the records of a zone and returns them in dns.RR format.
func (c *axfrddnsProvider) FetchZoneRecords(domain string) ([]dns.RR, error) {

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
			err := msg.Error.Error()
			if err == "dns: bad xfr rcode: 9" {
				err = "NOT AUTH (9)"
			}
			return nil, fmt.Errorf("[Error] AXFRDDNS: nameserver refused to transfer the zone: %s", err)
		}
		rawRecords = append(rawRecords, msg.RR...)
	}
	return rawRecords, nil

}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *axfrddnsProvider) GetZoneRecords(domain string) (models.Records, error) {

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
				foundDNSSecRecords.SetLabel(dnssecDummyLabel, domain)
				err = foundDNSSecRecords.SetTargetTXT(dnssecDummyTxt)
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
		// The SOA is sent two times: as the first and the last record
		// See section 2.2 of RFC5936
		foundRecords = foundRecords[:len(foundRecords)-1]
	}

	if foundDNSSecRecords != nil {
		foundRecords = append(foundRecords, foundDNSSecRecords)
	}

	return foundRecords, nil

}

// GetDomainCorrections returns a list of corrections to update a domain.
func (c *axfrddnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
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
			last.Name == dnssecDummyLabel &&
			len(last.TxtStrings) == 1 &&
			last.TxtStrings[0] == dnssecDummyTxt {
			hasDnssecRecords = true
			foundRecords = foundRecords[0:(len(foundRecords) - 1)]
		}
	}

	// TODO(tlim): This check should be done on all providers. Move to the global validation code.
	if dc.AutoDNSSEC == "on" && !hasDnssecRecords {
		fmt.Printf("Warning: AUTODNSSEC is enabled, but no DNSKEY or RRSIG record was found in the AXFR answer!\n")
	}
	if dc.AutoDNSSEC == "off" && hasDnssecRecords {
		fmt.Printf("Warning: AUTODNSSEC is disabled, but DNSKEY or RRSIG records were found in the AXFR answer!\n")
	}

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, create, del, mod, err := differ.IncrementalDiff(foundRecords)
	if err != nil {
		return nil, err
	}

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

					// An RFC2136-compliant server must silently ignore an
					// update that inserts a non-CNAME RRset when a CNAME RR
					// with the same name is present in the zone (and
					// vice-versa). Therefore we prefer to first remove records
					// and then insert new ones.
					//
					// Compliant servers must also silently ignore an update
					// that removes the last NS record of a zone. Therefore we
					// don't want to remove all NS records before inserting a
					// new one. For the particular case of NS record, we prefer
					// to insert new records before ot remove old ones.
					//
					// This remarks does not apply for "modified" NS records, as
					// updates are processed one-by-one.
					//
					// This provider does not allow modifying the TTL of an NS
					// record in a zone that defines only one NS. That would
					// would require removing the single NS record, before
					// adding the new one. But who does that anyway?

					update := new(dns.Msg)
					update.SetUpdate(dc.Name + ".")
					update.Id = uint16(c.rand.Intn(math.MaxUint16))
					for _, c := range create {
						if c.Desired.Type == "NS" {
							update.Insert([]dns.RR{c.Desired.ToRR()})
						}
					}
					for _, c := range del {
						update.Remove([]dns.RR{c.Existing.ToRR()})
					}
					for _, c := range mod {
						update.Remove([]dns.RR{c.Existing.ToRR()})
						update.Insert([]dns.RR{c.Desired.ToRR()})
					}
					for _, c := range create {
						if c.Desired.Type != "NS" {
							update.Insert([]dns.RR{c.Desired.ToRR()})
						}
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
						return fmt.Errorf("[Error] AXFRDDNS: nameserver refused to update the zone: %s (%d)",
							dns.RcodeToString[msg.MsgHdr.Rcode],
							msg.MsgHdr.Rcode)
					}

					return nil
				},
			})
	}
	return corrections, nil
}
