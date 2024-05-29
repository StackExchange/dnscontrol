package axfrddns

/*

axfrddns -
  Fetch the zone with an AXFR request (RFC5936) to a given primary master, and
  push Dynamic DNS updates (RFC2136) to the same server.

  Both the AXFR request and the updates might be authentificated with
  a TSIG.

*/

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/fatih/color"
	"github.com/miekg/dns"
)

const (
	dnsTimeout       = 30 * time.Second
	dnssecDummyLabel = "__dnssec"
	dnssecDummyTxt   = "Domain has DNSSec records, not displayed here."
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Can("Just warn when DNSSEC is requested but no RRSIG is found in the AXFR or warn when DNSSEC is not requested but RRSIG are found in the AXFR."),
	providers.CanGetZones:            providers.Cannot(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDHCID:            providers.Can(),
	providers.CanUseHTTPS:            providers.Can(),
	providers.CanUseLOC:              providers.Unimplemented(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseSVCB:             providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// axfrddnsProvider stores the client info for the provider.
type axfrddnsProvider struct {
	rand                *rand.Rand
	master              string
	updateMode          string
	transferServer      string
	transferMode        string
	nameservers         []*models.Nameserver
	transferKey         *Key
	updateKey           *Key
	hasDnssecRecords    bool
	serverHasBuggyCNAME bool
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
	if config["update-mode"] != "" {
		switch config["update-mode"] {
		case "tcp",
			"tcp-tls":
			api.updateMode = config["update-mode"]
		case "udp":
			api.updateMode = ""
		default:
			printer.Printf("[Warning] AXFRDDNS: Unknown update-mode in `creds.json` (%s)\n", config["update-mode"])
		}
	} else {
		api.updateMode = ""
	}
	if config["transfer-mode"] != "" {
		switch config["transfer-mode"] {
		case "tcp",
			"tcp-tls":
			api.transferMode = config["transfer-mode"]
		default:
			printer.Printf("[Warning] AXFRDDNS: Unknown transfer-mode in `creds.json` (%s)\n", config["transfer-mode"])
		}
	} else {
		api.transferMode = "tcp"
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
	if config["transfer-server"] != "" {
		api.transferServer = config["transfer-server"]
		if !strings.Contains(api.transferServer, ":") {
			api.transferServer = api.transferServer + ":53"
		}
	} else {
		api.transferServer = api.master
	}
	api.updateKey, err = readKey(config["update-key"], "update-key")
	if err != nil {
		return nil, err
	}
	api.transferKey, err = readKey(config["transfer-key"], "transfer-key")
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimSpace(config["buggy-cname"])) {
	case "yes", "true":
		api.serverHasBuggyCNAME = true
	default:
		api.serverHasBuggyCNAME = false
	}
	for key := range config {
		switch key {
		case "master",
			"nameservers",
			"update-key",
			"transfer-key",
			"transfer-server",
			"update-mode",
			"transfer-mode",
			"domain",
			"TYPE":
			continue
		default:
			printer.Printf("[Warning] AXFRDDNS: unknown key in `creds.json` (%s)\n", key)
		}
	}
	return api, err
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   initAxfrDdns,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("AXFRDDNS", fns, features)
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
	id := dns.CanonicalName(arr[1])
	return &Key{algo: algo, id: id, secret: arr[2]}, nil
}

// GetNameservers returns the nameservers for a domain.
func (c *axfrddnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return c.nameservers, nil
}

func (c *axfrddnsProvider) getAxfrConnection() (*dns.Transfer, error) {
	var con net.Conn = nil
	var err error = nil
	if c.transferMode == "tcp-tls" {
		con, err = tls.Dial("tcp", c.transferServer, &tls.Config{})
	} else {
		con, err = net.Dial("tcp", c.transferServer)
	}
	if err != nil {
		return nil, err
	}
	dnscon := &dns.Conn{Conn: con}
	transfer := &dns.Transfer{Conn: dnscon}
	return transfer, nil
}

// FetchZoneRecords gets the records of a zone and returns them in dns.RR format.
func (c *axfrddnsProvider) FetchZoneRecords(domain string) ([]dns.RR, error) {
	transfer, err := c.getAxfrConnection()
	if err != nil {
		return nil, err
	}
	transfer.DialTimeout = dnsTimeout
	transfer.ReadTimeout = dnsTimeout

	request := new(dns.Msg)
	request.SetAxfr(domain + ".")

	if c.transferKey != nil {
		transfer.TsigSecret =
			map[string]string{c.transferKey.id: c.transferKey.secret}
		request.SetTsig(c.transferKey.id, c.transferKey.algo, 300, time.Now().Unix())
		if c.transferKey.algo == dns.HmacMD5 {
			transfer.TsigProvider = md5Provider(c.transferKey.secret)
		}
	}

	envelope, err := transfer.In(request, c.transferServer)
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
			return nil, fmt.Errorf("[Error] AXFRDDNS: nameserver refused to transfer the zone %s: %s", domain, err)
		}
		rawRecords = append(rawRecords, msg.RR...)
	}
	return rawRecords, nil

}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *axfrddnsProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {

	rawRecords, err := c.FetchZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	var foundDNSSecRecords *models.RecordConfig
	foundRecords := models.Records{}
	for _, rr := range rawRecords {
		switch rr.Header().Rrtype {
		case dns.TypeRRSIG,
			dns.TypeDNSKEY,
			dns.TypeCDNSKEY,
			dns.TypeCDS,
			dns.TypeNSEC,
			dns.TypeNSEC3,
			dns.TypeNSEC3PARAM,
			65534:
			// Ignoring DNSSec RRs, but replacing it with a single
			// "TXT" placeholder
			// Also ignoring spurious TYPE65534, see:
			// https://bind9-users.isc.narkive.com/zX29ay0j/rndc-signing-list-not-working#post2
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
			rec, err := models.RRtoRC(rr, domain)
			if err != nil {
				return nil, err
			}
			foundRecords = append(foundRecords, &rec)
		}
	}

	if len(foundRecords) >= 1 && foundRecords[len(foundRecords)-1].Type == "SOA" {
		// The SOA is sent two times: as the first and the last record
		// See section 2.2 of RFC5936. We remove the later one.
		foundRecords = foundRecords[:len(foundRecords)-1]
	}

	if foundDNSSecRecords != nil {
		foundRecords = append(foundRecords, foundDNSSecRecords)
	}

	c.hasDnssecRecords = false
	if len(foundRecords) >= 1 {
		last := foundRecords[len(foundRecords)-1]
		if last.Type == "TXT" &&
			last.Name == dnssecDummyLabel &&
			last.GetTargetTXTSegmentCount() == 1 &&
			last.GetTargetTXTSegmented()[0] == dnssecDummyTxt {
			c.hasDnssecRecords = true
			foundRecords = foundRecords[0:(len(foundRecords) - 1)]
		}
	}

	return foundRecords, nil

}

// BuildCorrection return a Correction for a given set of DDNS update and the corresponding message.
func (c *axfrddnsProvider) BuildCorrection(dc *models.DomainConfig, msgs []string, update *dns.Msg) *models.Correction {
	if update == nil {
		return &models.Correction{
			Msg: fmt.Sprintf("DDNS UPDATES to '%s' (primary master: '%s'). Changes:\n%s", dc.Name, c.master, strings.Join(msgs, "\n")),
		}
	}
	return &models.Correction{
		Msg: fmt.Sprintf("DDNS UPDATES to '%s' (primary master: '%s'). Changes:\n%s", dc.Name, c.master, strings.Join(msgs, "\n")),
		F: func() error {

			client := new(dns.Client)
			client.Net = c.updateMode
			client.Timeout = dnsTimeout
			if c.updateKey != nil {
				client.TsigSecret =
					map[string]string{c.updateKey.id: c.updateKey.secret}
				update.SetTsig(c.updateKey.id, c.updateKey.algo, 300, time.Now().Unix())
				if c.updateKey.algo == dns.HmacMD5 {
					client.TsigProvider = md5Provider(c.updateKey.secret)
				}
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
	}
}

// hasDeletionForName returns true if there exist a corrections for [name] which is a deletion
func hasDeletionForName(changes diff2.ChangeList, name string) bool {
	for _, change := range changes {
		switch change.Type {
		case diff2.DELETE:
			if change.Old[0].Name == name {
				return true
			}
		}
	}
	return false
}

// hasNSDeletion returns true if there exist a correction that deletes or changes an NS record
func hasNSDeletion(changes diff2.ChangeList) bool {
	for _, change := range changes {
		switch change.Type {
		case diff2.CHANGE:
			if change.Old[0].Type == "NS" && change.Old[0].Name == "@" {
				return true
			}
		case diff2.DELETE:
			if change.Old[0].Type == "NS" && change.Old[0].Name == "@" {
				return true
			}
		case diff2.CREATE:
		case diff2.REPORT:
		}
	}
	return false
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (c *axfrddnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, foundRecords models.Records) ([]*models.Correction, error) {
	// Ignoring the SOA, others providers don't manage it either.
	if len(foundRecords) >= 1 && foundRecords[0].Type == "SOA" {
		foundRecords = foundRecords[1:]
	}

	// TODO(tlim): This check should be done on all providers. Move to the global validation code.
	if dc.AutoDNSSEC == "on" && !c.hasDnssecRecords {
		printer.Printf("Warning: AUTODNSSEC is enabled, but no DNSKEY or RRSIG record was found in the AXFR answer!\n")
	}
	if dc.AutoDNSSEC == "off" && c.hasDnssecRecords {
		printer.Printf("Warning: AUTODNSSEC is disabled, but DNSKEY or RRSIG records were found in the AXFR answer!\n")
	}

	// An RFC2136-compliant server must silently ignore an
	// update that inserts a non-CNAME RRset when a CNAME RR
	// with the same name is present in the zone (and
	// vice-versa). Therefore we prefer to first remove records
	// and then insert new ones.
	//
	// Compliant servers must also silently ignore an update
	// that removes the last NS record of a zone. Therefore we
	// don't want to remove all NS records before inserting a
	// new one. Then, when an update want to change a NS record,
	// we first insert a dummy NS record that we will remove
	// at the end of the batched update.

	var msgs []string
	var reports []string
	var msgs2 []string
	update := new(dns.Msg)
	update.SetUpdate(dc.Name + ".")
	update.Id = uint16(c.rand.Intn(math.MaxUint16))
	update2 := new(dns.Msg)
	update2.SetUpdate(dc.Name + ".")
	update2.Id = uint16(c.rand.Intn(math.MaxUint16))
	hasTwoCorrections := false

	dummyNs1, err := dns.NewRR(dc.Name + ". IN NS 255.255.255.255")
	if err != nil {
		return nil, err
	}
	dummyNs2, err := dns.NewRR(dc.Name + ". IN NS 255.255.255.255")
	if err != nil {
		return nil, err
	}

	changes, err := diff2.ByRecord(foundRecords, dc, nil)
	if err != nil {
		return nil, err
	}
	if changes == nil {
		return nil, nil
	}

	// A DNS server should silently ignore a DDNS update that removes
	// the last NS record of a zone. Since modifying a record is
	// implemented by successively a deletion of the old record and an
	// insertion of the new one, then modifying all the NS record of a
	// zone might will fail (even if the the deletion and insertion
	// are grouped in a single batched update).
	//
	// To avoid this case, we will first insert a dummy NS record,
	// that will be removed at the end of the batched updates. This
	// record needs to inserted only when all NS records are touched
	// The current implementation insert this dummy record as soon as
	// a NS record is deleted or changed.
	hasNSDeletion := hasNSDeletion(changes)

	if hasNSDeletion {
		update.Insert([]dns.RR{dummyNs1})
	}

	for _, change := range changes {
		switch change.Type {
		case diff2.DELETE:
			msgs = append(msgs, change.Msgs[0])
			update.Remove([]dns.RR{change.Old[0].ToRR()})
		case diff2.CREATE:
			if c.serverHasBuggyCNAME &&
				change.New[0].Type == "CNAME" &&
				hasDeletionForName(changes, change.New[0].Name) {
				hasTwoCorrections = true
				msgs2 = append(msgs2, change.Msgs[0])
				update2.Insert([]dns.RR{change.New[0].ToRR()})
			} else {
				msgs = append(msgs, change.Msgs[0])
				update.Insert([]dns.RR{change.New[0].ToRR()})
			}
		case diff2.CHANGE:
			if c.serverHasBuggyCNAME && change.New[0].Type == "CNAME" {
				msgs = append(msgs, change.Msgs[0]+color.RedString(" (delete)"))
				update.Remove([]dns.RR{change.Old[0].ToRR()})
				hasTwoCorrections = true
				msgs2 = append(msgs2, change.Msgs[0]+color.GreenString(" (create)"))
				update2.Insert([]dns.RR{change.New[0].ToRR()})
			} else {
				msgs = append(msgs, change.Msgs[0])
				update.Remove([]dns.RR{change.Old[0].ToRR()})
				update.Insert([]dns.RR{change.New[0].ToRR()})
			}
		case diff2.REPORT:
			reports = append(reports, change.Msgs...)
		}
	}

	if hasNSDeletion {
		update.Remove([]dns.RR{dummyNs2})
	}

	returnValue := []*models.Correction{}

	if len(msgs) > 0 {
		returnValue = append(returnValue, c.BuildCorrection(dc, msgs, update))
	}
	if hasTwoCorrections && len(msgs2) > 0 {
		returnValue = append(returnValue, c.BuildCorrection(dc, msgs2, update2))
	}
	if len(reports) > 0 {
		returnValue = append(returnValue, c.BuildCorrection(dc, reports, nil))
	}
	return returnValue, nil
}
