package gandi5

/*

Gandi API v5 LiveDNS provider:

Documentation: https://api.gandi.net/docs/
Endpoint: https://api.gandi.net/

Settings from `creds.json`:
   - apikey
   - sharing_id (optional)

*/

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/miekg/dns/dnsutil"
	gandi "github.com/tiramiseb/go-gandi-livedns"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/pkg/errors"
)

// Section 1: Register this provider in the system.

// init registers the provider to dnscontrol.
func init() {
	providers.RegisterDomainServiceProviderType("GANDI_V5", newDsp, features)
	providers.RegisterRegistrarType("GANDI_V5", newReg)
}

// features declares which features and options are available.
var features = providers.DocumentationNotes{
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("Can only manage domains registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// Section 2: Define the API client.

// gandiApi is the API handle used to store any client-related state.
type gandiApi struct {
	apikey    string
	sharingid string
}

// newDsp generates a DNS Service Provider client handle.
func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newHelper(conf, metadata)
}

// newReg generates a Registrar Provider client handle.
func newReg(conf map[string]string) (providers.Registrar, error) {
	return newHelper(conf, nil)
}

// newHelper generates a handle handle.
func newHelper(m map[string]string, metadata json.RawMessage) (*gandiApi, error) {
	api := &gandiApi{}
	api.apikey = m["apikey"]
	api.sharingid = m["sharing_id"]
	if api.apikey == "" {
		return nil, errors.Errorf("missing Gandi apikey")
	}

	return api, nil
}

// Section 3: DSP-related functions

func (client *gandiApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	existing, err := client.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	models.PostProcessRecords(existing)
	clean := PrepFoundRecods(existing)
	PrepDesiredRecords(dc)
	return client.GenerateDomainCorrections(dc, clean)
}

// GetZoneRecords gathers the DNS records and converts them to our
// standard format.
func (client *gandiApi) GetZoneRecords(domain string) ([]*models.RecordConfig, error) {
	g := gandi.New(client.apikey, client.sharingid)

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

func PrepFoundRecods(recs []*models.RecordConfig) []*models.RecordConfig {
	// If there are records that need to be modified, removed, etc. we
	// do it here.  This is usually empty.
	return recs
}

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

// debugRecords prints a list of RecordConfig.
func debugRecords(note string, recs []*models.RecordConfig) {
	fmt.Println("DEBUG:", note)
	for k, v := range recs {
		fmt.Printf("   %v: %v %v %v %v\n", k, v.GetLabel(), v.Type, v.TTL, v.GetTargetCombined())
	}
}

// debugKeyMap prints the data returned by ChangedGroups.
func debugKeyMap(note string, m map[models.RecordKey][]string) {
	fmt.Println("DEBUG:", note)

	// Extract the keys
	var keys []models.RecordKey
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i].NameFQDN == keys[j].NameFQDN {
			return keys[i].Type < keys[j].Type
		}
		return keys[i].NameFQDN < keys[j].NameFQDN
	})

	// Pretty print the map:
	for _, k := range keys {
		fmt.Printf("   %v %v:\n", k.Type, k.NameFQDN)
		for _, s := range m[k] {
			fmt.Printf("      -- %q\n", s)
		}
	}
}

// gatherAffectedLabels takes the output of diff.ChangedGroups and
// returns a the labels that are mentioned, as well as what messages
// to output when updating that label.  The two return values probably
// could be combined, but this works so I'm not going to mess with it.
func gatherAffectedLabels(groups map[models.RecordKey][]string) (affected map[string]bool, msgs map[string][]string) {
	affected = map[string]bool{}
	msgs = map[string][]string{}
	for k, v := range groups {
		affected[k.NameFQDN] = true
		msgs[k.NameFQDN] = append(msgs[k.NameFQDN], v...)
	}
	return affected, msgs
}

// gatherDesiredAtEachLabel extracts the RecordConfigs related to the
// labels mentioned in map m.
func gatherDesiredAtEachLabel(m map[string]bool, recs models.Records) (work map[string][]*models.RecordConfig) {
	//	desiredRecords = gatherDesiredAtEachLabel(affectedLabels, dc.Records)
	work = map[string][]*models.RecordConfig{}
	for label, _ := range m {
		for _, rec := range recs {
			if rec.GetLabelFQDN() == label {
				work[label] = append(work[label], rec)
			}
		}
	}
	return work
}

// gatherLabels returns a map of the labels mentioned in all the
// records.
func gatherLabels(recs []*models.RecordConfig) (labs map[string]bool) {
	labs = map[string]bool{}
	for _, r := range recs {
		labs[r.GetLabelFQDN()] = true
	}
	return labs
}

func (client *gandiApi) GenerateDomainCorrections(dc *models.DomainConfig, existing []*models.RecordConfig) ([]*models.Correction, error) {
	//debugRecords("GenDC input", existing)

	var corrections = []*models.Correction{}

	// diff
	differ := diff.New(dc)
	keysToUpdate := differ.ChangedGroups(existing)
	//debugKeyMap("GenDC diff", keysToUpdate)
	if len(keysToUpdate) == 0 {
		return nil, nil
	}

	// ChangedGroups returns data grouped by label:RType tuples. We need
	// that information listed by label.  Extract the info we need in
	// the format we need for later.
	affectedLabels, msgsForLabel := gatherAffectedLabels(keysToUpdate)
	desiredRecords := gatherDesiredAtEachLabel(affectedLabels, dc.Records)
	doesLabelExist := gatherLabels(existing)

	g := gandi.New(client.apikey, client.sharingid)

	// For any key with an update, delete or replace those records.
	for label, _ := range affectedLabels {
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
						//fmt.Printf("DEBUG: DeleteDomainRecords(%q, %q)\n", domain, shortname)
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
							//fmt.Printf("DEBUG: g.ChangeDomainRecordsWithName(%q, %q, %q)\n", domain, shortname, ns)
							res, err := g.ChangeDomainRecordsWithName(domain, shortname, ns)
							if err != nil {
								//fmt.Printf("DEBUG: g.res=%+v\n", res)
								return errors.Wrapf(err, "%+v", res)
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
								//fmt.Printf("DEBUG: CreateDomainRecord(%q, %q, %q, %q, %q)\n", domain, label, rtype, ttl, values)
								res, err := g.CreateDomainRecord(domain, shortname, rtype, ttl, values)
								if err != nil {
									//fmt.Printf("DEBUG: res=%+v\n", res)
									return errors.Wrapf(err, "%+v", res)
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

// Section 3: Registrat-related functions

// GetNameservers returns a list of nameservers for domain.
func (client *gandiApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	g := gandi.New(client.apikey, client.sharingid)
	nameservers, err := g.GetDomainNS(domain)
	if err != nil {
		return nil, err
	}
	return models.StringsToNameservers(nameservers), nil
}

// GetRegistrarCorrections returns a list of corrections for this registrar.
func (client *gandiApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	g := gandi.New(client.apikey, client.sharingid)

	nss, err := client.GetNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	existingNs := models.NameserversToStrings(nss)

	sort.Strings(existingNs)
	existing := strings.Join(existingNs, ",")

	desiredNs := []string{}
	for _, d := range dc.Nameservers {
		desiredNs = append(desiredNs, d.Name)
	}
	sort.Strings(desiredNs)
	desired := strings.Join(desiredNs, ",")

	if existing != desired {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", existing, desired),
				F: func() (err error) {
					err = g.UpdateDomainNS(dc.Name, desiredNs)
					return
				}},
		}, nil
	}
	return nil, nil
}
