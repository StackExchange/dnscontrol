package gandi5

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/miekg/dns/dnsutil"
	gandi "github.com/tiramiseb/go-gandi-livedns"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/pkg/errors"
)

/*

Gandi API v5 LiveDNS provider:

Documentation: https://api.gandi.net/docs/
Endpoint: https://api.gandi.net/

Info required in `creds.json`:
   - FILL_IN

*/

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
	//	domainIndex map[string]int64 // Map of domainname to index
	//	nameservers map[string][]*models.Nameserver
	//	ZoneId      int64
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

// GetNameservers returns a list of nameservers for domain.
func (client *gandiApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	g := gandi.New(client.apikey, client.sharingid)
	nameservers, err := g.GetDomainNS(domain)
	if err != nil {
		return nil, err
	}
	return models.StringsToNameservers(nameservers), nil
}

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
	records, err := g.ListDomainRecords(domain)
	if err != nil {
		return nil, err
	}

	//	fmt.Printf("RECORDS: %+v\n", records)

	// convert to dnscontrol RecordConfig format
	existingRecords := []*models.RecordConfig{}
	for _, rr := range records {
		existingRecords = append(existingRecords, nativeToRecords(rr, domain)...)
	}

	return existingRecords, nil
}

func PrepFoundRecods(recs []*models.RecordConfig) []*models.RecordConfig {
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

func (client *gandiApi) GenerateDomainCorrections(dc *models.DomainConfig, existing []*models.RecordConfig) ([]*models.Correction, error) {

	var corrections = []*models.Correction{}

	// diff
	differ := diff.New(dc)
	keysToUpdate := differ.ChangedGroups(existing)

	if len(keysToUpdate) == 0 {
		return nil, nil
	}

	// Gather all the labels that need to be updated, along with the
	// records to store there.
	labelsToUpdate := map[string][]*models.RecordConfig{}
	var msgsForLabel = map[string][]string{}
	for k, el := range keysToUpdate {
		label := dnsutil.TrimDomainName(k.NameFQDN, dc.Name)
		labelsToUpdate[label] = nil
		msgsForLabel[label] = append(msgsForLabel[label], el...)
		for _, rc := range dc.Records {
			if rc.GetLabel() == label {
				labelsToUpdate[label] = append(labelsToUpdate[label], rc)
			}
		}
	}

	// Make a map of what keys exist. This is used later to determine if
	// labelsToUpdate are on new or existing keys.
	existingLabels := map[string]bool{}
	for _, r := range dc.Records {
		existingLabels[r.GetLabel()] = true
	}

	g := gandi.New(client.apikey, client.sharingid)

	// For any key with an update, delete or replace those records.
	for label, recs := range labelsToUpdate {
		if len(recs) == 0 {
			// No records matching this key?  This can only mean that all
			// the records were deleted. Delete them.

			//domain, label, rtype := dc.Name, dnsutil.TrimDomainName(k.NameFQDN, dc.Name), k.Type
			domain := dc.Name
			msgs := strings.Join(msgsForLabel[label], "\n")
			corrections = append(corrections,
				&models.Correction{
					Msg: msgs,
					F: func() error {
						fmt.Printf("DEBUG: DeleteDomainRecords(%q, %q)\n", domain, label)
						err := g.DeleteDomainRecords(domain, label)
						if err != nil {
							return err
						}
						return nil
					},
				})

		} else {
			// Replace all the records at a label with our new records.

			// Generate the new data in Gandi's format.
			ns := recordsToNative(recs, dc.Name)
			//fmt.Printf("NS=%+v\n", ns)

			domain := dc.Name
			label := recs[0].GetLabel()
			labelfqdn := recs[0].GetLabelFQDN()

			if _, ok := existingLabels[label]; !ok {
				// First time putting data on this label.

				// We have to create the label one rtype at a time.
				for _, n := range ns {
					rtype := n.RrsetType
					ttl := n.RrsetTTL
					values := n.RrsetValues
					//fmt.Printf("VALUES1 = %q\n", values)

					key := models.RecordKey{NameFQDN: labelfqdn, Type: rtype}
					msgs := strings.Join(keysToUpdate[key], "\n")

					corrections = append(corrections,
						&models.Correction{
							Msg: msgs,
							F: func() error {
								fmt.Printf("DEBUG: CreateDomainRecord(%q, %q, %q, %q, %q)\n", domain, label, rtype, ttl, values)
								res, err := g.CreateDomainRecord(domain, label, rtype, ttl, values)
								if err != nil {
									fmt.Printf("DEBUG: res=%+v\n", res)
									return errors.Wrapf(err, "%+v", res)
								}
								return nil
							},
						})
				}

			} else {
				// Records exist for this label. Replace them with what we have.
				msgs := strings.Join(msgsForLabel[label], "\n")
				//fmt.Printf("VALUES2 = %+v\n", ns)
				corrections = append(corrections,
					&models.Correction{
						Msg: msgs,
						F: func() error {
							fmt.Printf("DEBUG: g.ChangeDomainRecordsWithName(%q, %q, %q)\n", domain, label, ns)
							res, err := g.ChangeDomainRecordsWithName(domain, label, ns)
							if err != nil {
								fmt.Printf("DEBUG: g.res=%+v\n", res)
								return errors.Wrapf(err, "%+v", res)
							}
							return nil
						},
					})

			}
		}

	}
	return corrections, nil
}

/*


	If a key has zero updates, it is a delete.
	How are we going to do deletes?
		DeleteDomainRecords(fqdn, name string) (err error)

	If it has any changes (it is non-zero), then we are lazy and just replace all the records for that label.
	We could be fancy and update the exact records, but this is

	ChangeDomainRecordsWithName

*/

//type gandiRecord struct {
//	gandirecord.RecordInfo
//}

// func (c *gandiApi) getDomainInfo(domain string) (*gandidomain.DomainInfo, error) {
// 	if err := c.fetchDomainList(); err != nil {
// 		return nil, err
// 	}
// 	_, ok := c.domainIndex[domain]
// 	if !ok {
// 		return nil, errors.Errorf("%s not listed in zones for gandi account", domain)
// 	}
// 	return c.fetchDomainInfo(domain)
// }
//
// // GetNameservers returns the nameservers for domain.
// func (c *gandiApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
// 	domaininfo, err := c.getDomainInfo(domain)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ns := []*models.Nameserver{}
// 	for _, nsname := range domaininfo.Nameservers {
// 		ns = append(ns, &models.Nameserver{Name: nsname})
// 	}
// 	return ns, nil
// }
//
// // GetDomainCorrections returns a list of corrections recommended for this domain.
// func (c *gandiApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
// 	dc.Punycode()
// 	domaininfo, err := c.getDomainInfo(dc.Name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	foundRecords, err := c.getZoneRecords(domaininfo.ZoneId, dc.Name)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	expectedRecordSets := make([]gandirecord.RecordSet, 0, len(dc.Records))
// 	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))
// 	for _, rec := range dc.Records {
// 		if rec.TTL < 300 {
// 			printer.Warnf("Gandi does not support ttls < 300. Setting %s from %d to 300\n", rec.GetLabelFQDN(), rec.TTL)
// 			rec.TTL = 300
// 		}
// 		if rec.TTL > 2592000 {
// 			return nil, errors.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
// 		}
// 		if rec.Type == "TXT" {
// 			rec.SetTarget("\"" + rec.GetTargetField() + "\"") // FIXME(tlim): Should do proper quoting.
// 		}
// 		if rec.Type == "NS" && rec.GetLabel() == "@" {
// 			if !strings.HasSuffix(rec.GetTargetField(), ".gandi.net.") {
// 				printer.Warnf("Gandi does not support changing apex NS records. %s will not be added.\n", rec.GetTargetField())
// 			}
// 			continue
// 		}
// 		rs := gandirecord.RecordSet{
// 			"type":  rec.Type,
// 			"name":  rec.GetLabel(),
// 			"value": rec.GetTargetCombined(),
// 			"ttl":   rec.TTL,
// 		}
// 		expectedRecordSets = append(expectedRecordSets, rs)
// 		recordsToKeep = append(recordsToKeep, rec)
// 	}
// 	dc.Records = recordsToKeep
//
// 	// Normalize
// 	models.PostProcessRecords(foundRecords)
//
// 	differ := diff.New(dc)
// 	_, create, del, mod := differ.IncrementalDiff(foundRecords)
//
// 	// Print a list of changes. Generate an actual change that is the zone
// 	changes := false
// 	desc := ""
// 	for _, i := range create {
// 		changes = true
// 		desc += "\n" + i.String()
// 	}
// 	for _, i := range del {
// 		changes = true
// 		desc += "\n" + i.String()
// 	}
// 	for _, i := range mod {
// 		changes = true
// 		desc += "\n" + i.String()
// 	}
//
// 	msg := fmt.Sprintf("GENERATE_ZONE: %s (%d records)%s", dc.Name, len(dc.Records), desc)
// 	corrections := []*models.Correction{}
// 	if changes {
// 		corrections = append(corrections,
// 			&models.Correction{
// 				Msg: msg,
// 				F: func() error {
// 					printer.Printf("CREATING ZONE: %v\n", dc.Name)
// 					return c.createGandiZone(dc.Name, domaininfo.ZoneId, expectedRecordSets)
// 				},
// 			})
// 	}
//
// 	return corrections, nil
// }

// GetRegistrarCorrections returns a list of corrections for this registrar.
func (c *gandiApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	// 	domaininfo, err := c.getDomainInfo(dc.Name)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	sort.Strings(domaininfo.Nameservers)
	// 	found := strings.Join(domaininfo.Nameservers, ",")
	// 	desiredNs := []string{}
	// 	for _, d := range dc.Nameservers {
	// 		desiredNs = append(desiredNs, d.Name)
	// 	}
	// 	sort.Strings(desiredNs)
	// 	desired := strings.Join(desiredNs, ",")
	// 	if found != desired {
	// 		return []*models.Correction{
	// 			{
	// 				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", found, desired),
	// 				F: func() (err error) {
	// 					_, err = c.setDomainNameservers(dc.Name, desiredNs)
	// 					return
	// 				}},
	// 		}, nil
	// 	}
	return nil, nil
}
