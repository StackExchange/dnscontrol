package gandiv5

/*

Gandi API V5 LiveDNS provider:

Documentation: https://api.gandi.net/docs/
Endpoint: https://api.gandi.net/

Settings from `creds.json`:
   - apikey
   - sharing_id (optional)

*/

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/go-gandi/go-gandi"
	"github.com/go-gandi/go-gandi/config"
	"github.com/miekg/dns/dnsutil"
)

// Section 1: Register this provider in the system.

// init registers the provider to dnscontrol.
func init() {
	fns := providers.DspFuncs{
		Initializer:   newDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("GANDI_V5", fns, features)
	providers.RegisterRegistrarType("GANDI_V5", newReg)
}

// features declares which features and options are available.
var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can("Only on the bare domain. Otherwise CNAME will be substituted"),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot("Only supports DS records at the apex"),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Cannot("Can only manage domains registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// DNSSEC: platform supports it, but it doesn't fit our GetDomainCorrections
// model, so deferring for now.

// Section 2: Define the API client.

// gandiv5Provider is the gandiv5Provider handle used to store any client-related state.
type gandiv5Provider struct {
	apikey    string
	sharingid string
	debug     bool
}

// newDsp generates a DNS Service Provider client handle.
func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newHelper(conf, metadata)
}

// newReg generates a Registrar Provider client handle.
func newReg(conf map[string]string) (providers.Registrar, error) {
	return newHelper(conf, nil)
}

// newHelper generates a handle.
func newHelper(m map[string]string, _ json.RawMessage) (*gandiv5Provider, error) {
	api := &gandiv5Provider{}
	api.apikey = m["apikey"]
	if api.apikey == "" {
		return nil, fmt.Errorf("missing Gandi apikey")
	}
	api.sharingid = m["sharing_id"]
	debug, err := strconv.ParseBool(os.Getenv("GANDI_V5_DEBUG"))
	if err == nil {
		api.debug = debug
	}

	return api, nil
}

// Section 3: Domain Service Provider (DSP) related functions

// // ListZones lists the zones on this account.
// This no longer works. Until we can figure out why, we're removing this
// feature for Gandi.
// func (client *gandiv5Provider) ListZones() ([]string, error) {
// 	g := gandi.NewLiveDNSClient(config.Config{
// 		APIKey:    client.apikey,
// 		SharingID: client.sharingid,
// 		Debug:     client.debug,
// 	})

// 	listResp, err := g.ListDomains()
// 	if err != nil {
// 		return nil, err
// 	}

// 	zones := make([]string, len(listResp))
// 	fmt.Printf("DEBUG: HERE START\n")
// 	for i, zone := range listResp {
// 	fmt.Printf("DEBUG: HERE %d: %v\n", i, zone.FQDN)
// 		zone := zone
// 		zones[i] = zone.FQDN
// 	}
// 	fmt.Printf("DEBUG: HERE END\n")
// 	return zones, nil
// }

// GetZoneRecords gathers the DNS records and converts them to
// dnscontrol's format.
func (client *gandiv5Provider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	g := gandi.NewLiveDNSClient(config.Config{
		APIKey:    client.apikey,
		SharingID: client.sharingid,
		Debug:     client.debug,
	})

	// Get all the existing records:
	records, err := g.GetDomainRecords(domain)
	if err != nil {
		return nil, err
	}

	// Convert them to DNScontrol's native format:
	existingRecords := []*models.RecordConfig{}
	for _, rr := range records {
		rrs, err := nativeToRecords(rr, domain)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rrs...)
	}

	return existingRecords, nil
}

// PrepDesiredRecords munges any records to best suit this provider.
func PrepDesiredRecords(dc *models.DomainConfig) {
	// Sort through the dc.Records, eliminate any that can't be
	// supported; modify any that need adjustments to work with the
	// provider.  We try to do minimal changes otherwise it gets
	// confusing.

	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" && rec.Name != "@" {
			// GANDI only permits aliases on a naked domain.
			// Therefore, we change this to a CNAME.
			rec.Type = "CNAME"
		}
		if rec.TTL < 300 {
			printer.Warnf("Gandi does not support ttls < 300. Setting %s from %d to 300\n", rec.GetLabelFQDN(), rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			printer.Warnf("Gandi does not support ttls > 30 days. Setting %s from %d to 2592000\n", rec.GetLabelFQDN(), rec.TTL)
			rec.TTL = 2592000
		}
		if rec.Type == "NS" && rec.GetLabel() == "@" {
			if !strings.HasSuffix(rec.GetTargetField(), ".gandi.net.") {
				printer.Warnf("Gandi does not support changing apex NS records. Ignoring %s\n", rec.GetTargetField())
			}
			continue
		}
		recordsToKeep = append(recordsToKeep, rec)
	}
	dc.Records = recordsToKeep
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (client *gandiv5Provider) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	var corrections []*models.Correction
	if client.debug {
		debugRecords("GenDC input", existing)
	}

	PrepDesiredRecords(dc)

	g := gandi.NewLiveDNSClient(config.Config{
		APIKey:    client.apikey,
		SharingID: client.sharingid,
		Debug:     client.debug,
	})

	// Gandi is a "ByLabel" API with the odd exception that changes must be
	// done one label:rtype at a time.
	instructions, err := diff2.ByLabel(existing, dc, nil)
	if err != nil {
		return nil, err
	}
	for _, inst := range instructions {
		switch inst.Type {

		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: inst.MsgsJoined})

		case diff2.CREATE:
			// We have to create the label one rtype at a time.
			// In other words, this is a ByRecordSet API for creation, even though
			// this is very ByLabel()-ish for everything else.
			natives := recordsToNative(inst.New, dc.Name)
			for _, n := range natives {
				label := inst.Key.NameFQDN
				rtype := n.RrsetType
				domain := dc.Name
				shortname := dnsutil.TrimDomainName(label, dc.Name)
				ttl := n.RrsetTTL
				values := n.RrsetValues
				key := models.RecordKey{NameFQDN: label, Type: rtype}
				msg := strings.Join(inst.MsgsByKey[key], "\n")
				corrections = append(corrections,
					&models.Correction{
						Msg: msg,
						F: func() error {
							res, err := g.CreateDomainRecord(domain, shortname, rtype, ttl, values)
							if err != nil {
								return fmt.Errorf("%+v: %w", res, err)
							}
							return nil
						},
					})
			}

		case diff2.CHANGE:
			msgs := strings.Join(inst.Msgs, "\n")
			domain := dc.Name
			label := inst.Key.NameFQDN
			shortname := dnsutil.TrimDomainName(label, dc.Name)
			ns := recordsToNative(inst.New, dc.Name)
			corrections = append(corrections,
				&models.Correction{
					Msg: msgs,
					F: func() error {
						res, err := g.UpdateDomainRecordsByName(domain, shortname, ns)
						if err != nil {
							return fmt.Errorf("%+v: %w", res, err)
						}
						return nil
					},
				})

		case diff2.DELETE:
			msgs := strings.Join(inst.Msgs, "\n")
			domain := dc.Name
			label := inst.Key.NameFQDN
			shortname := dnsutil.TrimDomainName(label, dc.Name)
			corrections = append(corrections,
				&models.Correction{
					Msg: msgs,
					F: func() error {
						err := g.DeleteDomainRecordsByName(domain, shortname)
						if err != nil {
							return err
						}
						return nil
					},
				})

		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", inst.Type))
		}

	}

	return corrections, nil
}

// debugRecords prints a list of RecordConfig.
func debugRecords(note string, recs []*models.RecordConfig) {
	printer.Debugf(note)
	for k, v := range recs {
		printer.Printf("   %v: %v %v %v %v\n", k, v.GetLabel(), v.Type, v.TTL, v.GetTargetDebug())
	}
}

// Section 3: Registrar-related functions

// GetNameservers returns a list of nameservers for domain.
func (client *gandiv5Provider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	g := gandi.NewLiveDNSClient(config.Config{
		APIKey:    client.apikey,
		SharingID: client.sharingid,
		Debug:     client.debug,
	})
	nameservers, err := g.GetDomainNS(domain)
	if err != nil {
		return nil, err
	}
	return models.ToNameservers(nameservers)
}

// GetRegistrarCorrections returns a list of corrections for this registrar.
func (client *gandiv5Provider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	gd := gandi.NewDomainClient(config.Config{
		APIKey:    client.apikey,
		SharingID: client.sharingid,
		Debug:     client.debug,
	})

	existingNs, err := gd.GetNameServers(dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(existingNs)
	existing := strings.Join(existingNs, ",")

	desiredNs := models.NameserversToStrings(dc.Nameservers)
	sort.Strings(desiredNs)
	desired := strings.Join(desiredNs, ",")

	if existing != desired {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", existing, desired),
				F: func() (err error) {
					err = gd.UpdateNameServers(dc.Name, desiredNs)
					return
				}},
		}, nil
	}
	return nil, nil
}
