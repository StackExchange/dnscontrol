package gandi5

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

	"github.com/miekg/dns/dnsutil"
	gandi "github.com/tiramiseb/go-gandi"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/printer"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/diff"
)

// Section 1: Register this provider in the system.

// init registers the provider to dnscontrol.
func init() {
	providers.RegisterDomainServiceProviderType("GANDI_V5", newDsp, features)
	providers.RegisterRegistrarType("GANDI_V5", newReg)
}

// features declares which features and options are available.
var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("Can only manage domains registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

// DNSSEC: platform supports it, but it doesn't fit our GetDomainCorrections
// model, so deferring for now.

// Section 2: Define the API client.

// api is the api handle used to store any client-related state.
type api struct {
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
func newHelper(m map[string]string, metadata json.RawMessage) (*api, error) {
	api := &api{}
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

// NB(tal): To future-proof your code, all new providers should
// implement GetDomainCorrections exactly as you see here
// (byte-for-byte the same). In 3.0
// we plan on using just the individual calls to GetZoneRecords,
// PostProcessRecords, and so on.
//
// Currently every provider does things differently, which prevents
// us from doing things like using GetZoneRecords() of a provider
// to make convertzone work with all providers.

// GetDomainCorrections get the current and existing records,
// post-process them, and generate corrections.
func (client *api) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	existing, err := client.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	models.PostProcessRecords(existing)
	clean := PrepFoundRecords(existing)
	PrepDesiredRecords(dc)
	return client.GenerateDomainCorrections(dc, clean)
}

// GetZoneRecords gathers the DNS records and converts them to
// dnscontrol's format.
func (client *api) GetZoneRecords(domain string) (models.Records, error) {
	g := gandi.NewLiveDNSClient(client.apikey, gandi.Config{SharingID: client.sharingid, Debug: client.debug})

	// Get all the existing records:
	records, err := g.ListDomainRecords(domain)
	if err != nil {
		return nil, err
	}

	// Convert them to DNScontrol's native format:
	existingRecords := []*models.RecordConfig{}
	for _, rr := range records {
		existingRecords = append(existingRecords, nativeToRecords(rr, domain)...)
	}

	return existingRecords, nil
}

// PrepFoundRecords munges any records to make them compatible with
// this provider. Usually this is a no-op.
func PrepFoundRecords(recs models.Records) models.Records {
	// If there are records that need to be modified, removed, etc. we
	// do it here.  Usually this is a no-op.
	return recs
}

// PrepDesiredRecords munges any records to best suit this provider.
func PrepDesiredRecords(dc *models.DomainConfig) {
	// Sort through the dc.Records, eliminate any that can't be
	// supported; modify any that need adjustments to work with the
	// provider.  We try to do minimal changes otherwise it gets
	// confusing.

	dc.Punycode()

	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.TTL < 300 {
			printer.Warnf("Gandi does not support ttls < 300. Setting %s from %d to 300\n", rec.GetLabelFQDN(), rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			printer.Warnf("Gandi does not support ttls > 30 days. Setting %s from %d to 2592000\n", rec.GetLabelFQDN(), rec.TTL)
			rec.TTL = 2592000
		}
		if rec.Type == "TXT" {
			rec.SetTarget("\"" + rec.GetTargetField() + "\"") // FIXME(tlim): Should do proper quoting.
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

// GenerateDomainCorrections takes the desired and existing records
// and produces a Correction list.  The correction list is simply
// a list of functions to call to actually make the desired
// correction, and a message to output to the user when the change is
// made.
func (client *api) GenerateDomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	if client.debug {
		debugRecords("GenDC input", existing)
	}

	var corrections = []*models.Correction{}

	// diff existing vs. current.
	differ := diff.New(dc)
	keysToUpdate := differ.ChangedGroups(existing)
	if client.debug {
		diff.DebugKeyMapMap("GenDC diff", keysToUpdate)
	}
	if len(keysToUpdate) == 0 {
		return nil, nil
	}

	// Regroup data by FQDN.  ChangedGroups returns data grouped by label:RType tuples.
	affectedLabels, msgsForLabel := gatherAffectedLabels(keysToUpdate)
	_, desiredRecords := dc.Records.GroupedByFQDN()
	doesLabelExist := existing.FQDNMap()

	g := gandi.NewLiveDNSClient(client.apikey, gandi.Config{SharingID: client.sharingid, Debug: client.debug})

	// For any key with an update, delete or replace those records.
	for label := range affectedLabels {
		if len(desiredRecords[label]) == 0 {
			// No records matching this key?  This can only mean that all
			// the records were deleted. Delete them.

			msgs := strings.Join(msgsForLabel[label], "\n")
			domain := dc.Name
			shortname := dnsutil.TrimDomainName(label, dc.Name)
			corrections = append(corrections,
				&models.Correction{
					Msg: msgs,
					F: func() error {
						err := g.DeleteDomainRecords(domain, shortname)
						if err != nil {
							return err
						}
						return nil
					},
				})

		} else {
			// Replace all the records at a label with our new records.

			// Generate the new data in Gandi's format.
			ns := recordsToNative(desiredRecords[label], dc.Name)

			if doesLabelExist[label] {
				// Records exist for this label. Replace them with what we have.

				msg := strings.Join(msgsForLabel[label], "\n")
				domain := dc.Name
				shortname := dnsutil.TrimDomainName(label, dc.Name)
				corrections = append(corrections,
					&models.Correction{
						Msg: msg,
						F: func() error {
							res, err := g.UpdateDomainRecordsByName(domain, shortname, ns)
							if err != nil {
								return fmt.Errorf("%+v: %w", res, err)
							}
							return nil
						},
					})

			} else {
				// First time putting data on this label. Create it.

				// We have to create the label one rtype at a time.
				for _, n := range ns {
					msg := strings.Join(msgsForLabel[label], "\n")
					domain := dc.Name
					shortname := dnsutil.TrimDomainName(label, dc.Name)
					rtype := n.RrsetType
					ttl := n.RrsetTTL
					values := n.RrsetValues
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
			}
		}
	}

	return corrections, nil
}

// debugRecords prints a list of RecordConfig.
func debugRecords(note string, recs []*models.RecordConfig) {
	fmt.Println("DEBUG:", note)
	for k, v := range recs {
		fmt.Printf("   %v: %v %v %v %v\n", k, v.GetLabel(), v.Type, v.TTL, v.GetTargetCombined())
	}
}

// gatherAffectedLabels takes the output of diff.ChangedGroups and
// regroups it by FQDN of the label, not by Key. It also returns
// a list of all the FQDNs.
func gatherAffectedLabels(groups map[models.RecordKey][]string) (labels map[string]bool, msgs map[string][]string) {
	labels = map[string]bool{}
	msgs = map[string][]string{}
	for k, v := range groups {
		labels[k.NameFQDN] = true
		msgs[k.NameFQDN] = append(msgs[k.NameFQDN], v...)
	}
	return labels, msgs
}

// Section 3: Registrar-related functions

// GetNameservers returns a list of nameservers for domain.
func (client *api) GetNameservers(domain string) ([]*models.Nameserver, error) {
	g := gandi.NewLiveDNSClient(client.apikey, gandi.Config{SharingID: client.sharingid, Debug: client.debug})
	nameservers, err := g.GetDomainNS(domain)
	if err != nil {
		return nil, err
	}
	return models.ToNameservers(nameservers)
}

// GetRegistrarCorrections returns a list of corrections for this registrar.
func (client *api) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	gd := gandi.NewDomainClient(client.apikey, gandi.Config{SharingID: client.sharingid, Debug: client.debug})

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
